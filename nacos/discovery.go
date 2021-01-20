/**
 * @Time : 2021/1/11 8:16 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/loadbalance"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/pojo"
	"github.com/MassAdobe/go-gateway/utils"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	Instances              sync.Map // 实体的调用地址容器
	RequestTmzMap          sync.Map // 请求次数记录
	GrayScaleInstances     sync.Map // 灰度发布的调用地址容器
	GrayScaleRequestTmzMap sync.Map // 灰度发布的次数记录
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 3:16 下午
 * @Description: nacos服务注册发现
**/
func NacosDiscovery() {
	if pojo.InitConf.NacosDiscovery {
		// 创建动态配置客户端
		var namingClientErr error
		// 创建服务发现客户端
		namingClient, namingClientErr = clients.CreateNamingClient(map[string]interface{}{
			constants.NACOS_SERVER_CONFIGS_MARK: serverCs,
			constants.NACOS_CLIENT_CONFIG_MARK:  clientC,
		})
		if nil != namingClientErr {
			logs.Lg.Error("nacos服务注册与发现", namingClientErr, logs.Desc("创建服务发现客户端失败"))
			os.Exit(1)
		}
		logs.Lg.Debug("nacos服务注册与发现", logs.Desc("创建服务发现客户端成功"))
		if ip, err := utils.ExternalIP(); err != nil {
			logs.Lg.Error("nacos服务注册与发现", err, logs.Desc("nacos获取当前机器IP失败"))
			os.Exit(1)
		} else {
			pojo.CurIp = ip.String() // 赋值当前宿主IP
			success, namingErr := namingClient.RegisterInstance(vo.RegisterInstanceParam{
				Ip:          pojo.CurIp,
				Port:        InitConfiguration.Serve.Port,
				ServiceName: InitConfiguration.Serve.ServerName,
				Weight:      InitConfiguration.Serve.Weight,
				Enable:      true,
				Healthy:     true,
				Ephemeral:   true,
				Metadata:    map[string]string{constants.NACOS_REGIST_IDC_MARK: constants.NACOS_REGIST_IDC_INNER, constants.NACOS_REGIST_TIMESTAMP_MARK: utils.RtnCurTime()},
				ClusterName: constants.NACOS_DISCOVERY_CLUSTER_NAME, // 默认值DEFAULT
				GroupName:   pojo.InitConf.NacosGroup,               // 默认值DEFAULT_GROUP
			})
			if !success || nil != namingErr {
				logs.Lg.Error("nacos服务注册与发现", namingErr, logs.Desc("nacos注册服务失败"))
				os.Exit(1)
			}
		}
		logs.Lg.Debug("nacos服务注册与发现", logs.Desc("服务注册成功"))
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 4:57 下午
 * @Description: nacos注销服务
**/
func NacosDeregister() {
	if pojo.InitConf.NacosDiscovery {
		success, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          pojo.CurIp,
			Port:        InitConfiguration.Serve.Port,
			ServiceName: InitConfiguration.Serve.ServerName,
			Ephemeral:   true,
			Cluster:     constants.NACOS_DISCOVERY_CLUSTER_NAME, // 默认值DEFAULT
			GroupName:   pojo.InitConf.NacosGroup,               // 默认值DEFAULT_GROUP
		})
		if !success || nil != err {
			logs.Lg.Error("nacos服务注册与发现", err, logs.Desc("nacos注销服务失败"))
			os.Exit(1)
		}
		logs.Lg.Debug("nacos服务注册与发现", logs.Desc("nacos服务注销成功"))
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/11 8:58 下午
 * @Description: 获取实例（初次获取）
**/
func InitNacosGetInstances() {
	if len(InitConfiguration.Routers.LoadBalance) != 0 {
		loadbalance.Lb = &loadbalance.LoadBalance{Type: strings.ToLower(InitConfiguration.Routers.LoadBalance)}
	}
	if 0 == InitConfiguration.Routers.RefreshTmz { // 设置刷新次数参数
		RefreshTmz = 50 // 默认50次
	} else {
		RefreshTmz = InitConfiguration.Routers.RefreshTmz
	}
	if len(InitConfiguration.Routers.Version) == 0 {
		os.Exit(1)
	}
	Version = strings.ToLower(InitConfiguration.Routers.Version) // 获取路由的版本信息
	for k, v := range InitConfiguration.Routers.Services {
		instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
			ServiceName: k,
			GroupName:   v, // 默认值DEFAULT_GROUP
			Clusters:    []string{Version},
		})
		if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
			logs.Lg.Error("获取实例", err, logs.Desc("获取实例失败"))
		}
		RequestTmzMap.Store(k, 0)        // 添加调用次数
		loadbalance.Lb.Round.Store(k, 0) // 新增次数记录
		if len(instances) == 0 {         // 如果列表为空
			Instances.Store(k, nil)
		} else { // 列表不为空
			urls := make([]*url.URL, 0)
			for _, val := range instances {
				urls = append(urls, &url.URL{
					Scheme: constants.NACOS_SCHEMA,
					Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
				})
			}
			Instances.Store(k, urls)
		}
		if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
			loadbalance.Lb.Round.Store(k, 0)
		}
	}

	// 如果当前的灰度发布是开的状态 并且是自研方式 统计服务
	if InitConfiguration.GrayScale.Open && strings.ToLower(InitConfiguration.Routers.LoadBalance) != constants.NACOS_CONFIGURATION_MARK {
		for k, v := range InitConfiguration.Routers.Services {
			instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
				ServiceName: k,
				GroupName:   v, // 默认值DEFAULT_GROUP
				Clusters:    []string{InitConfiguration.GrayScale.Version},
			})
			if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
				logs.Lg.Error("获取实例", err, logs.Desc("获取实例失败(灰度)"))
			}
			GrayScaleRequestTmzMap.Store(k, 0)        // 添加调用次数(灰度)
			loadbalance.Lb.GrayScaleRound.Store(k, 0) // 新增次数记录(灰度)
			if len(instances) == 0 {                  // 如果列表为空(灰度)
				GrayScaleInstances.Store(k, nil)
			} else { // 列表不为空
				urls := make([]*url.URL, 0)
				for _, val := range instances {
					urls = append(urls, &url.URL{
						Scheme: constants.NACOS_SCHEMA,
						Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
					})
				}
				GrayScaleInstances.Store(k, urls)
			}
			if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
				loadbalance.Lb.GrayScaleRound.Store(k, 0)
			}
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 10:00 上午
 * @Description: 请求中获取实例
**/
func NacosGetInstances(serviceName string) {
	logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("获取的服务实例(请求中): %s", serviceName)))
	if _, okay := Instances.Load(serviceName); !okay { // 如果当前服务不存在 nacos中没有配置
		logs.Lg.Error("获取实例", errors.New("current service has not been configured in nacos"), logs.Desc(fmt.Sprintf("当前服务: %s没有在nacos的路由中配置", serviceName)))
		return
	}
	instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: serviceName,
		GroupName:   InitConfiguration.Routers.Services[serviceName], // 默认值DEFAULT_GROUP
		Clusters:    []string{Version},
	})
	if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
		logs.Lg.Error("获取实例", err, logs.Desc(fmt.Sprintf("获取实例失败(请求中)，服务: %s", serviceName)))
		panic(errs.NewError(errs.ErrNacosGetInstanceCode))
	}
	if len(instances) == 0 { // 如果列表为空
		logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("注册中心服务列表为空，服务: %s", serviceName)))
		Instances.Store(serviceName, nil)
	} else { // 列表不为空
		logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("注册中心服务列表不为空，服务：%s", serviceName)))
		urls := make([]*url.URL, 0)
		for _, val := range instances {
			urls = append(urls, &url.URL{
				Scheme: constants.NACOS_SCHEMA,
				Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
			})
		}
		Instances.Store(serviceName, urls)
	}
	if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
		logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("当前配置为自研强轮训，设置轮训参数，服务: %s", serviceName)))
		loadbalance.Lb.Round.Store(serviceName, 0)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/14 10:51 上午
 * @Description: 请求中获取实例(灰度)
**/
func NacosGetGrayScaleInstances(serviceName string) {
	// 如果当前的灰度发布是开的状态 并且是自研方式 统计服务
	logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("获取的服务实例(请求中，灰度): %s", serviceName)))
	if PuGrayScale.Open && strings.ToLower(loadbalance.Lb.Type) != constants.NACOS_CONFIGURATION_MARK {
		if _, okay := Instances.Load(serviceName); !okay { // 如果当前服务不存在 nacos中没有配置
			logs.Lg.Error("获取实例", errors.New("current service has not been configured in nacos"), logs.Desc(fmt.Sprintf("当前服务: %s没有在nacos的路由中配置(灰度)", serviceName)))
			return
		}
		instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
			ServiceName: serviceName,
			GroupName:   InitConfiguration.Routers.Services[serviceName], // 默认值DEFAULT_GROUP
			Clusters:    []string{PuGrayScale.Version},
		})
		if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
			logs.Lg.Error("获取实例", err, logs.Desc(fmt.Sprintf("获取实例失败(请求中，灰度)，服务: %s", serviceName)))
			panic(errs.NewError(errs.ErrNacosGetInstanceCode))
		}
		if len(instances) == 0 { // 如果列表为空
			logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("注册中心服务列表为空，灰度，服务: %s", serviceName)))
			GrayScaleInstances.Store(serviceName, nil)
		} else { // 列表不为空
			logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("注册中心服务列表不为空，灰度，服务：%s", serviceName)))
			urls := make([]*url.URL, 0)
			for _, val := range instances {
				urls = append(urls, &url.URL{
					Scheme: constants.NACOS_SCHEMA,
					Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
				})
			}
			GrayScaleInstances.Store(serviceName, urls)
		}
		if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
			logs.Lg.Debug("获取实例", logs.Desc(fmt.Sprintf("当前配置为自研强轮训，设置轮训参数，灰度，服务: %s", serviceName)))
			loadbalance.Lb.GrayScaleRound.Store(serviceName, 0)
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 11:08 上午
 * @Description: 监听获取实例
**/
func NacosGetInstancesListener(profile *InitNacosConfiguration) {
	logs.Lg.Debug("nacos配置文件监听", logs.Desc("路由配置变更"))
	if len(profile.Routers.Version) == 0 {
		logs.Lg.Error("nacos配置文件监听", errors.New("router version is nil"), logs.Desc("路由版本信息为空"))
		return
	}
	Version = strings.ToLower(profile.Routers.Version)
	if len(profile.Routers.LoadBalance) != 0 {
		loadbalance.Lb.Type = strings.ToLower(profile.Routers.LoadBalance)
	}
	RefreshTmz = profile.Routers.RefreshTmz // 设置刷新次数参数
	logs.Lg.Debug("nacos配置文件监听", logs.Desc("设置路由刷新次数参数"))
	// 先删除不存在的
	Instances.Range(func(key, value interface{}) bool {
		if _, okay := profile.Routers.Services[key.(string)]; !okay {
			Instances.Delete(key)     // 删除服务记录数据
			RequestTmzMap.Delete(key) // 删除调用次数
			if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
				loadbalance.Lb.Round.Delete(key) // 删除轮训数据
				logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("删除路由: %s的配置", key)))
			}
		}
		return true
	})
	// 插入新的
	for k, v := range profile.Routers.Services {
		instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
			ServiceName: k,
			GroupName:   v, // 默认值DEFAULT_GROUP
			Clusters:    []string{Version},
		})
		if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
			logs.Lg.Error("nacos配置文件监听", err, logs.Desc(fmt.Sprintf("获取路由实例失败(nacos监听)，服务: %s", k)))
			return
		}
		RequestTmzMap.Store(k, 0)        // 添加调用次数
		loadbalance.Lb.Round.Store(k, 0) // 新增次数记录
		// 如果列表为空
		if len(instances) == 0 {
			logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("注册中心服务列表为空，服务: %s", k)))
			Instances.Store(k, nil)
		} else { // 列表不为空
			urls := make([]*url.URL, 0)
			for _, val := range instances {
				urls = append(urls, &url.URL{
					Scheme: constants.NACOS_SCHEMA,
					Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
				})
			}
			Instances.Store(k, urls)
			logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("注册中心服务列表不为空，服务：%s", k)))
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/14 10:53 上午
 * @Description: 监听获取实例(灰度)
**/
func NacosGetGrayScaleInstancesListener(profile *InitNacosConfiguration) {
	// 如果当前的灰度发布是开的状态 并且是自研方式 统计服务 如果是开启的状态
	if PuGrayScale.Open && strings.ToLower(loadbalance.Lb.Type) != "nacos" {
		// 先删除不存在的
		GrayScaleInstances.Range(func(key, value interface{}) bool {
			if _, okay := profile.Routers.Services[key.(string)]; !okay {
				GrayScaleInstances.Delete(key)     // 删除服务记录数据
				GrayScaleRequestTmzMap.Delete(key) // 删除调用次数
				if strings.ToLower(loadbalance.Lb.Type) == constants.LOAD_BALANCE_ROUND {
					loadbalance.Lb.GrayScaleRound.Delete(key) // 删除轮训数据
					logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("删除灰度路由: %s的配置", key)))
				}
			}
			return true
		})
		// 插入新的
		for k, v := range profile.Routers.Services {
			instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
				ServiceName: k,
				GroupName:   v, // 默认值DEFAULT_GROUP
				Clusters:    []string{PuGrayScale.Version},
			})
			if err != nil && err.Error() != constants.INSTANCE_LIST_EMPTY {
				logs.Lg.Error("nacos配置文件监听", err, logs.Desc(fmt.Sprintf("获取路由实例失败(nacos监听)，灰度，服务: %s", k)))
				return
			}
			GrayScaleRequestTmzMap.Store(k, 0)        // 添加调用次数
			loadbalance.Lb.GrayScaleRound.Store(k, 0) // 新增次数记录
			// 如果列表为空
			if len(instances) == 0 {
				logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("注册中心服务列表为空，灰度，服务: %s", k)))
				GrayScaleInstances.Store(k, nil)
			} else { // 列表不为空
				urls := make([]*url.URL, 0)
				for _, val := range instances {
					urls = append(urls, &url.URL{
						Scheme: constants.NACOS_SCHEMA,
						Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
					})
				}
				GrayScaleInstances.Store(k, urls)
				logs.Lg.Debug("nacos配置文件监听", logs.Desc(fmt.Sprintf("注册中心服务列表不为空，灰度，服务：%s", k)))
			}
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/18 2:32 下午
 * @Description: 获取服务调用参数
**/
func NacosGetServer(serviceName, groupName, clusterName string) (instance *model.Instance, err error) {
	instance, err = namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   groupName,             // 默认值DEFAULT_GROUP
		Clusters:    []string{clusterName}, // 默认值DEFAULT
	})
	if err != nil {
		logs.Lg.Error("nacos服务注册与发现", err, logs.Desc("获取服务失败"))
		instance = nil
		return
	}
	logs.Lg.Debug("nacos服务注册与发现", logs.Desc(fmt.Sprintf("获取服务成功: %v", instance)))
	return
}

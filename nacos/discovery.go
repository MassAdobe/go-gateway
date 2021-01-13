/**
 * @Time : 2021/1/11 8:16 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"fmt"
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

const (
	INSTANCE_LIST_EMPTY = "instance list is empty!" // 列表为空错误
	DEFAULT_SCHEMA      = "http"                    // 固定请求方式
)

var (
	Instances     sync.Map // 实体的调用地址容器
	RequestTmzMap sync.Map // 请求次数记录
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
			"serverConfigs": serverCs,
			"clientConfig":  clientC,
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
				Metadata:    map[string]string{"idc": "shanghai", "timestamp": utils.RtnCurTime()},
				ClusterName: "DEFAULT",                // 默认值DEFAULT
				GroupName:   pojo.InitConf.NacosGroup, // 默认值DEFAULT_GROUP
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
			Cluster:     "DEFAULT",                // 默认值DEFAULT
			GroupName:   pojo.InitConf.NacosGroup, // 默认值DEFAULT_GROUP
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
	RefreshTmz = InitConfiguration.Routers.RefreshTmz // 设置刷新次数参数
	for k, v := range InitConfiguration.Routers.Services {
		instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
			ServiceName: k,
			GroupName:   v, // 默认值DEFAULT_GROUP
		})
		if err != nil && err.Error() != INSTANCE_LIST_EMPTY {
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
					Scheme: DEFAULT_SCHEMA,
					Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
				})
			}
			Instances.Store(k, urls)
		}
		if loadbalance.Lb.Type == loadbalance.LOAD_BALANCE_ROUND {
			loadbalance.Lb.Round.Store(k, 0)
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 10:00 上午
 * @Description: 请求中获取实例
**/
func NacosGetInstances(serviceName string) {
	instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: serviceName,
		GroupName:   InitConfiguration.Routers.Services[serviceName], // 默认值DEFAULT_GROUP
	})
	if err != nil && err.Error() != INSTANCE_LIST_EMPTY {
		logs.Lg.Error("获取实例", err, logs.Desc("获取实例失败(请求中)"))
	}
	if len(instances) == 0 { // 如果列表为空
		Instances.Store(serviceName, nil)
	} else { // 列表不为空
		urls := make([]*url.URL, 0)
		for _, val := range instances {
			urls = append(urls, &url.URL{
				Scheme: DEFAULT_SCHEMA,
				Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
			})
		}
		Instances.Store(serviceName, urls)
	}
	if loadbalance.Lb.Type == loadbalance.LOAD_BALANCE_ROUND {
		loadbalance.Lb.Round.Store(serviceName, 0)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 11:08 上午
 * @Description: 监听获取实例
**/
func NacosGetInstancesListener(profile *InitNacosConfiguration) {
	RefreshTmz = profile.Routers.RefreshTmz // 设置刷新次数参数
	// 先删除不存在的
	Instances.Range(func(key, value interface{}) bool {
		if _, okay := profile.Routers.Services[key.(string)]; !okay {
			Instances.Delete(key)     // 删除服务记录数据
			RequestTmzMap.Delete(key) // 删除调用次数
			if loadbalance.Lb.Type == loadbalance.LOAD_BALANCE_ROUND {
				loadbalance.Lb.Round.Delete(key) // 删除轮训数据
			}
		}
		return true
	})
	// 插入新的
	for k, v := range profile.Routers.Services {
		instances, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
			ServiceName: k,
			GroupName:   v, // 默认值DEFAULT_GROUP
		})
		if err != nil && err.Error() != INSTANCE_LIST_EMPTY {
			logs.Lg.Error("获取实例", err, logs.Desc("获取实例失败(nacos监听)"))
		}
		RequestTmzMap.Store(k, 0)        // 添加调用次数
		loadbalance.Lb.Round.Store(k, 0) // 新增次数记录
		// 如果列表为空
		if len(instances) == 0 {
			Instances.Store(k, nil)
		} else { // 列表不为空
			urls := make([]*url.URL, 0)
			for _, val := range instances {
				urls = append(urls, &url.URL{
					Scheme: DEFAULT_SCHEMA,
					Host:   fmt.Sprintf("%s:%d", val.Ip, val.Port),
				})
			}
			Instances.Store(k, urls)
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/18 2:32 下午
 * @Description: 获取服务调用参数
**/
func NacosGetServer(serviceName, groupName string) (instance *model.Instance, err error) {
	instance, err = namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   groupName,           // 默认值DEFAULT_GROUP
		Clusters:    []string{"DEFAULT"}, // 默认值DEFAULT
	})
	if err != nil {
		logs.Lg.Error("nacos服务注册与发现", err, logs.Desc("获取服务失败"))
		instance = nil
		return
	}
	logs.Lg.Debug("nacos服务注册与发现", logs.Desc(fmt.Sprintf("获取服务成功: %v", instance)))
	return
}

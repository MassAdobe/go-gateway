/**
 * @Time : 2021/1/11 8:04 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/pojo"
	"github.com/MassAdobe/go-gateway/utils"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

var (
	serverCs      []constant.ServerConfig     // nacos的server配置
	clientC       constant.ClientConfig       // nacos的client配置
	profileC      vo.ConfigParam              // nacos的配置
	configClient  config_client.IConfigClient // nacos服务配置中心client
	namingClient  naming_client.INamingClient // nacos服务注册与发现client
	NacosContent  string                      // nacos配置中心配置内容
	Version       string                      // 路由的版本信息
	RefreshTmz    int                         // 接口频次同步服务次数
	BlackList     map[string]bool             // 黑名单
	WhiteList     map[string]bool             // 白名单
	PuGrayScale   *GrayScale                  // 灰度发布
	ForceLoginOut map[int64]time.Time         // 强制下线配置表
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 9:58 上午
 * @Description: 初始化nacos
**/
func InitNacos() {
	if pojo.InitConf.NacosConfiguration || pojo.InitConf.NacosDiscovery {
		// 初始化nacos的server服务
		nacosIps := strings.Split(pojo.InitConf.NacosServerIps, ",")
		if 0 == len(nacosIps) {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos地址不能为空"))
			os.Exit(1)
		}
		if 0 == pojo.InitConf.NacosServerPort {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos端口号不能为空"))
			os.Exit(1)
		}
		for _, ip := range nacosIps {
			serverCs = append(serverCs, constant.ServerConfig{
				IpAddr:      ip,
				ContextPath: "/nacos",
				Port:        pojo.InitConf.NacosServerPort,
				Scheme:      "http",
			})
		}
		// 初始化nacos的client服务
		if 0 == len(pojo.InitConf.NacosClientNamespaceId) {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos命名空间不能为空"))
			os.Exit(1)
		}
		clientC = constant.ClientConfig{
			NamespaceId:         pojo.InitConf.NacosClientNamespaceId, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
			NotLoadCacheAtStart: true,
			LogDir:              "/tmp/nacos/log",
			CacheDir:            "/tmp/nacos/cache",
			RotateTime:          "1h",
			MaxAge:              3,
			LogLevel:            "debug",
		}
		if 0 == pojo.InitConf.NacosClientTimeoutMs {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos请求Nacos服务端的超时时间为空，默认为10000ms"))
			os.Exit(1)
		}
		clientC.TimeoutMs = pojo.InitConf.NacosClientTimeoutMs
	}
	if pojo.InitConf.NacosConfiguration {
		// 初始化nacos的获取配置服务
		profileC = vo.ConfigParam{
			DataId: pojo.InitConf.NacosDataId,
			Group:  pojo.InitConf.NacosGroup,
		}
		fmt.Println(fmt.Sprintf(`{"log_level":"INFO","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "初始化配置成功"))
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 2:50 下午
 * @Description: nacos配置中心
**/
func NacosConfiguration() {
	if pojo.InitConf.NacosConfiguration {
		// 创建动态配置客户端
		var configClientErr error
		configClient, configClientErr = clients.CreateConfigClient(map[string]interface{}{
			"serverConfigs": serverCs,
			"clientConfig":  clientC,
		})
		if nil != configClientErr {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos配置中心连接错误"))
			os.Exit(1)
		}
		// 获取配置
		var contentErr error
		if NacosContent, contentErr = configClient.GetConfig(profileC); contentErr != nil {
			fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "nacos配置中心获取配置错误"))
			os.Exit(1)
		}
		fmt.Println(fmt.Sprintf(`{"log_level":"INFO","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "获取配置成功"))
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/21 2:58 下午
 * @Description: 监听配置文件变化
**/
func ListenConfiguration() {
	if pojo.InitConf.NacosConfiguration {
		err := configClient.ListenConfig(vo.ConfigParam{
			DataId: pojo.InitConf.NacosDataId,
			Group:  pojo.InitConf.NacosGroup,
			OnChange: func(namespace, group, dataId, data string) {
				logs.Lg.Info("nacos配置文件监听", logs.Desc(fmt.Sprintf("groupId: %s, dataId: %s, data: %s", group, dataId, data)))
				// 修改日志级别
				profile := ReadNacosProfile(data)
				if strings.ToLower(pojo.InitConf.LogLevel) != strings.ToLower(profile.Log.Level) {
					logs.Lg.Debug("nacos配置文件监听", logs.Desc("日志级别修改"))
					switch strings.ToLower(profile.Log.Level) {
					case "debug":
						logs.Lg.Level.SetLevel(zap.DebugLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "info":
						logs.Lg.Level.SetLevel(zap.InfoLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "warn":
						logs.Lg.Level.SetLevel(zap.WarnLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "error":
						logs.Lg.Level.SetLevel(zap.ErrorLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "dpanic":
						logs.Lg.Level.SetLevel(zap.DPanicLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "panic":
						logs.Lg.Level.SetLevel(zap.PanicLevel)
						printModifiedLog(profile.Log.Level)
						break
					case "fatal":
						logs.Lg.Level.SetLevel(zap.FatalLevel)
						printModifiedLog(profile.Log.Level)
						break
					default:
						logs.Lg.Error("动态调整日志级别", errors.New("dynamic modified log level error"), logs.Desc("动态调整日志级别失败，日志级别字符不正确"))
						break
					}
				}
				// 动态修改灰度发布
				ModifiedGrayScale(profile)
				// 监听获取实例(灰度)
				NacosGetGrayScaleInstancesListener(profile)
				// 修改路由
				NacosGetInstancesListener(profile)
				// 动态修改黑白名单
				ModifiedBWList(profile)
				// 动态修改强制下线配置
				ModifiedForceLoginOut(profile)
			},
		})
		if err != nil {
			logs.Lg.Error("nacos配置文件监听", err, logs.Desc("设置nacos配置文件监听器失败"))
			os.Exit(1)
		}
		logs.Lg.Debug("nacos配置文件监听", logs.Desc("设置nacos配置文件监听器成功"))
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/21 6:02 下午
 * @Description: 输出动态修改日志级别日志，同时赋值新日志级别
**/
func printModifiedLog(current string) {
	logs.Lg.Info("动态调整日志级别",
		logs.Desc(fmt.Sprintf("，由级别 %s 调至 %s",
			strings.ToLower(pojo.InitConf.LogLevel), strings.ToLower(current))))
	pojo.InitConf.LogLevel = strings.ToLower(current)
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 8:11 下午
 * @Description: 灰度发布结构体
**/
type GrayScale struct {
	Open    bool            `yaml:"open"`    // 是否开启
	Version string          `yaml:"version"` // 需要灰度版本
	Type    string          `yaml:"type"`    // 种类：'userId':用户ID范围,'userList':用户列表,'ipList':IP列表
	List    map[string]bool `yaml:"list"`    // 配置列表
	Scope   *Scope          `yaml:"scope"`   // 范围
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 8:12 下午
 * @Description: 灰度发布内部范围结构体
**/
type Scope struct {
	Type string `yaml:"type"` // 种类：'great':大于,'less':小于
	Mark int64  `yaml:"mark"` // 值
}

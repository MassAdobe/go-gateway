/**
 * @Time : 2021/1/11 8:01 下午
 * @Author : MassAdobe
 * @Description: pojo
**/
package pojo

var (
	InitConf InitConfig // 初始化配置
	CurIp    string     // 当前宿主IP
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 2:06 下午
 * @Description: 初始化配置
**/
type InitConfig struct {
	NacosConfiguration     bool   `yaml:"NacosConfiguration"`     // 是否开启nacos配置中心
	NacosDiscovery         bool   `yaml:"NacosDiscovery"`         // 是否开启nacos服务注册于发现
	NacosServerIps         string `yaml:"NacosServerIps"`         // nacos地址
	NacosServerPort        uint64 `yaml:"NacosServerPort"`        // nacos端口号
	NacosClientNamespaceId string `yaml:"NacosClientNamespaceId"` // nacos命名空间
	NacosClientTimeoutMs   uint64 `yaml:"NacosClientTimeoutMs"`   // 请求Nacos服务端的超时时间，默认是10000ms
	NacosDataId            string `yaml:"NacosDataId"`            // nacos配置文件名称
	NacosGroup             string `yaml:"NacosGroup"`             // nacos配置组名称
	LogPath                string `yaml:"LogPath"`                // 日志输出路径(本地配置优先级最高)
	LogLevel               string `yaml:"LogLevel"`               // 日志级别(本地配置优先级最高)
}

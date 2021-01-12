/**
 * @Time : 2020/12/17 4:25 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import "fmt"

var (
	InitConfiguration InitNacosConfiguration // 初始化配置
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 4:26 下午
 * @Description: nacos配置文件配置
**/
type InitNacosConfiguration struct {
	Serve struct { // 服务配置
		Port       uint64  `yaml:"port"`        // 服务端口号
		ServerName string  `yaml:"server-name"` // 服务名
		Weight     float64 `yaml:"weight"`      // nacos中权重
	} `yaml:"serve"`

	Log struct { // 日志配置
		Path  string `yaml:"path"`  // 日志地址
		Level string `yaml:"level"` // 日志级别
	} `yaml:"log"`

	Routers struct {
		RefreshTmz  int               `yaml:"refresh-tmz"`  // 请求次数刷新服务
		LoadBalance string            `yaml:"load-balance"` // 负载均衡方法
		Services    map[string]string `yaml:"services"`     // 反向代理服务名和组名
	} `yaml:"routers"` // 路由
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/18 11:39 上午
 * @Description: 拼装请求主地址
**/
func RequestPath(path string) string {
	return fmt.Sprintf("/%s/%s", InitConfiguration.Serve.ServerName, path)
}

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

	BWList struct {
		BlackList []string `yaml:"black-list"` // 黑名单
		WhiteList []string `yaml:"white-list"` // 白名单
	} `yaml:"bw-list"` // 黑白名单

	GrayScale struct {
		Open    bool     `yaml:"open"`    // 是否开启
		Version string   `yaml:"version"` // 需要灰度版本
		Type    string   `yaml:"type"`    // 种类：'userId':用户ID范围,'userList':用户列表,'ipList':IP列表
		List    []string `yaml:"list"`    // 配置列表
		Scope   struct {
			Type string `yaml:"type"` // 种类：'great':大于,'less':小于
			Mark string `yaml:"mark"` // 值
		} `yaml:"scope"` // 范围
	} `yaml:"grayscale"` // 灰度发布
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/18 11:39 上午
 * @Description: 拼装请求主地址
**/
func RequestPath(path string) string {
	return fmt.Sprintf("/%s/%s", InitConfiguration.Serve.ServerName, path)
}

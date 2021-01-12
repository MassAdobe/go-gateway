/**
 * @Time : 2021/1/11 8:08 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"fmt"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/utils"
	"gopkg.in/yaml.v2"
	"os"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 4:24 下午
 * @Description: 处理首次nacos获取到的配置信息
**/
func InitNacosProfile() {
	if err := yaml.Unmarshal([]byte(NacosContent), &InitConfiguration); err != nil {
		fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", "未知", "读取nacos系统配置失败"))
		os.Exit(1)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/21 3:05 下午
 * @Description: 返回配置文件内容
**/
func ReadNacosProfile(content string) *InitNacosConfiguration {
	var NewInitConfiguration InitNacosConfiguration
	if err := yaml.Unmarshal([]byte(content), &NewInitConfiguration); err != nil {
		logs.Lg.Error("解析nacos配置", err, logs.Desc("解析nacos配置失败"))
	}
	return &NewInitConfiguration
}

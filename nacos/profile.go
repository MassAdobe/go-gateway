/**
 * @Time : 2021/1/11 8:08 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/utils"
	"gopkg.in/yaml.v2"
	"os"
	"sync"
)

var (
	mutex sync.RWMutex // 设置动态修改黑白名单的读写锁
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
	} else { // 配置黑白名单
		BlackList, WhiteList = make(map[string]bool), make(map[string]bool)
		// 初始化黑名单
		if len(InitConfiguration.BWList.BlackList) != 0 {
			if len(InitConfiguration.BWList.BlackList) != 0 {
				for _, val := range InitConfiguration.BWList.BlackList {
					if utils.CheckIp(val) { // 校验配置的IP地址是否正确
						BlackList[val] = true
					} else { // 不正确 不添加
						fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", InitConfiguration.Serve.ServerName, fmt.Sprintf("当前nacos配置的黑名单IP地址错误，IP为: %s", val)))
					}
				}
			}
		}
		// 初始化白名单
		if len(InitConfiguration.BWList.WhiteList) != 0 {
			if len(InitConfiguration.BWList.WhiteList) != 0 {
				for _, val := range InitConfiguration.BWList.WhiteList {
					if utils.CheckIp(val) { // 校验配置的IP地址是否正确
						WhiteList[val] = true
					} else { // 不正确 不添加
						fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", InitConfiguration.Serve.ServerName, fmt.Sprintf("当前nacos配置的白名单IP地址错误，IP为: %s", val)))
					}
				}
			}
		}
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

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 5:26 下午
 * @Description: 动态修改黑白名单
**/
func ModifiedBWList(profile *InitNacosConfiguration) {
	logs.Lg.Debug("动态修改黑白名单", logs.Desc("开始修改黑白名单"))
	mutex.RLock()
	defer mutex.RUnlock()
	// 初始化黑名单
	if len(profile.BWList.BlackList) != 0 {
		// 先删除
		for k := range BlackList {
			mark := true
			for _, val := range profile.BWList.BlackList {
				if k == val {
					mark = false
					break
				}
			}
			if mark {
				logs.Lg.Debug("动态修改黑白名单", logs.Desc(fmt.Sprintf("删除黑名单，IP为: %s", k)))
				delete(BlackList, k)
			}
		}
		// 再增加
		for _, val := range profile.BWList.BlackList {
			if utils.CheckIp(val) { // 校验配置的IP地址是否正确
				if _, okay := BlackList[val]; !okay {
					BlackList[val] = true
				}
			} else { // 不正确 不添加
				logs.Lg.Error("动态修改黑白名单", errors.New("create black list ip error"), logs.Desc(fmt.Sprintf("当前新增的黑名单IP地址错误，IP为: %s", val)))
			}
		}
	} else {
		logs.Lg.Debug("动态修改黑白名单", logs.Desc("当前没有配置黑名单"))
		BlackList = make(map[string]bool)
	}
	// 初始化白名单
	if len(profile.BWList.WhiteList) != 0 {
		// 先删除
		for k := range WhiteList {
			mark := true
			for _, val := range profile.BWList.WhiteList {
				if k == val {
					mark = false
					break
				}
			}
			if mark {
				logs.Lg.Debug("动态修改黑白名单", logs.Desc(fmt.Sprintf("删除白名单，IP为: %s", k)))
				delete(WhiteList, k)
			}
		}
		// 再增加
		for _, val := range profile.BWList.WhiteList {
			if _, okay := WhiteList[val]; !okay {
				if utils.CheckIp(val) { // 校验配置的IP地址是否正确
					WhiteList[val] = true
				}
			} else { // 不正确 不添加
				logs.Lg.Error("动态修改黑白名单", errors.New("create white list ip error"), logs.Desc(fmt.Sprintf("当前新增的白名单IP地址错误，IP为: %s", val)))
			}
		}
	} else {
		logs.Lg.Debug("动态修改黑白名单", logs.Desc("当前没有配置白名单"))
		WhiteList = make(map[string]bool)
	}
}

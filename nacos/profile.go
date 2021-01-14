/**
 * @Time : 2021/1/11 8:08 下午
 * @Author : MassAdobe
 * @Description: nacos
**/
package nacos

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/loadbalance"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/utils"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
	"sync"
)

const (
	GRAY_SCALE_USER_ID_TYPE   = "userscope" // 灰度发布种类：用户ID范围
	GRAY_SCALE_USER_LIST_TYPE = "userlist"  // 灰度发布种类：用户列表
	GRAY_SCALE_IP_LIST_TYPE   = "iplist"    // 灰度发布种类：IP列表
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
	} else {
		// 配置黑白名单
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
		// 配置灰度发布
		if InitConfiguration.GrayScale.Open && len(InitConfiguration.GrayScale.Version) != 0 && len(InitConfiguration.GrayScale.Type) != 0 {
			switch strings.ToLower(InitConfiguration.GrayScale.Type) {
			case GRAY_SCALE_USER_ID_TYPE: // 灰度发布种类：用户ID范围
				PuGrayScale = &GrayScale{
					Open:    InitConfiguration.GrayScale.Open,
					Version: strings.ToLower(InitConfiguration.GrayScale.Version),
					Type:    strings.ToLower(InitConfiguration.GrayScale.Type),
					Scope: &Scope{
						Type: strings.ToLower(InitConfiguration.GrayScale.Scope.Type),
						Mark: InitConfiguration.GrayScale.Scope.Mark,
					},
				}
				break
			case GRAY_SCALE_USER_LIST_TYPE: // 灰度发布种类：用户列表
				PuGrayScale = &GrayScale{
					Open:    InitConfiguration.GrayScale.Open,
					Version: strings.ToLower(InitConfiguration.GrayScale.Version),
					Type:    strings.ToLower(InitConfiguration.GrayScale.Type),
					List:    make(map[string]bool),
				}
				for _, val := range InitConfiguration.GrayScale.List {
					PuGrayScale.List[val] = true
				}
				break
			case GRAY_SCALE_IP_LIST_TYPE: // 灰度发布种类：IP列表
				PuGrayScale = &GrayScale{
					Open:    InitConfiguration.GrayScale.Open,
					Version: strings.ToLower(InitConfiguration.GrayScale.Version),
					Type:    strings.ToLower(InitConfiguration.GrayScale.Type),
					List:    make(map[string]bool),
				}
				for _, val := range InitConfiguration.GrayScale.List {
					if utils.CheckIp(val) {
						PuGrayScale.List[val] = true
						continue
					}
					fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", InitConfiguration.Serve.ServerName, fmt.Sprintf("当前开启的灰度发布配置的IP地址有误，IP: %s", val)))
				}
				break
			default:
				fmt.Println(fmt.Sprintf(`{"log_level":"ERROR","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "配置中心", InitConfiguration.Serve.ServerName, fmt.Sprintf("当前开启的灰度发布种类错误，种类为: %s", InitConfiguration.GrayScale.Type)))
				break
			}
		} else { // 关闭
			PuGrayScale = &GrayScale{
				Open: false,
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
			if _, okay := BlackList[val]; !okay {
				if utils.CheckIp(val) { // 校验配置的IP地址是否正确
					BlackList[val] = true
				} else { // 不正确 不添加
					logs.Lg.Error("动态修改黑白名单", errors.New("create black list ip error"), logs.Desc(fmt.Sprintf("当前新增的黑名单IP地址错误，IP为: %s", val)))
				}
				continue
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
				} else { // 不正确 不添加
					logs.Lg.Error("动态修改黑白名单", errors.New("create white list ip error"), logs.Desc(fmt.Sprintf("当前新增的白名单IP地址错误，IP为: %s", val)))
				}
				continue
			}
		}
	} else {
		logs.Lg.Debug("动态修改黑白名单", logs.Desc("当前没有配置白名单"))
		WhiteList = make(map[string]bool)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 8:27 下午
 * @Description: 动态修改灰度发布
**/
func ModifiedGrayScale(profile *InitNacosConfiguration) {
	logs.Lg.Debug("动态修改灰度发布", logs.Desc("开始修改灰度发布配置"))
	mutex.RLock()
	defer mutex.RUnlock()
	// 配置灰度发布
	if profile.GrayScale.Open { // 开启
		logs.Lg.Debug("动态修改灰度发布", logs.Desc("开启灰度发布"))
		if len(profile.GrayScale.Version) != 0 && len(profile.GrayScale.Type) != 0 {
			switch strings.ToLower(profile.GrayScale.Type) {
			case GRAY_SCALE_USER_ID_TYPE: // 灰度发布种类：用户ID范围
				PuGrayScale = &GrayScale{
					Open:    profile.GrayScale.Open,
					Version: strings.ToLower(profile.GrayScale.Version),
					Type:    strings.ToLower(profile.GrayScale.Type),
					Scope: &Scope{
						Type: strings.ToLower(profile.GrayScale.Scope.Type),
						Mark: profile.GrayScale.Scope.Mark,
					},
				}
				break
			case GRAY_SCALE_USER_LIST_TYPE: // 灰度发布种类：用户列表
				PuGrayScale = &GrayScale{
					Open:    profile.GrayScale.Open,
					Version: strings.ToLower(profile.GrayScale.Version),
					Type:    strings.ToLower(profile.GrayScale.Type),
					List:    make(map[string]bool),
				}
				for _, val := range profile.GrayScale.List {
					PuGrayScale.List[val] = true
				}
				break
			case GRAY_SCALE_IP_LIST_TYPE: // 灰度发布种类：IP列表
				PuGrayScale = &GrayScale{
					Open:    profile.GrayScale.Open,
					Version: strings.ToLower(profile.GrayScale.Version),
					Type:    strings.ToLower(profile.GrayScale.Type),
					List:    make(map[string]bool),
				}
				for _, val := range profile.GrayScale.List {
					if utils.CheckIp(val) {
						PuGrayScale.List[val] = true
						continue
					}
					logs.Lg.Error("动态修改灰度发布", errors.New("gray scale ip list error"), logs.Desc(fmt.Sprintf("修改的IP地址非法，IP: %s", val)))
				}
				break
			default:
				logs.Lg.Error("动态修改灰度发布", errors.New("gray scale type error"), logs.Desc(fmt.Sprintf("当前开启的灰度发布种类错误，种类为: %s", profile.GrayScale.Type)))
			}
		}
	} else { // 关闭
		logs.Lg.Debug("动态修改灰度发布", logs.Desc("关闭灰度发布"))
		PuGrayScale = &GrayScale{Open: false}
		// 如果没有配置，直接初始化所有灰度配置
		loadbalance.Lb.GrayScaleRound = sync.Map{}
		GrayScaleRequestTmzMap = sync.Map{}
		GrayScaleInstances = sync.Map{}
	}
}

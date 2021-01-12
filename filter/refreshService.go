/**
 * @Time : 2021/1/12 5:56 下午
 * @Author : MassAdobe
 * @Description: filter
**/
package filter

import (
	"github.com/MassAdobe/go-gateway/nacos"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 6:01 下午
 * @Description: 检查次数，如果超过了相关次数，重新获取服务信息
**/
func CheckTmz(serviceName string) {
	// 增加次数
	load, _ := nacos.RequestTmzMap.Load(serviceName)
	if load.(int) >= nacos.RefreshTmz { // 如果到达次数，那么需要重新获取该服务的发现列表
		nacos.RequestTmzMap.Store(serviceName, 0) // 到达次数，请求数归零
		nacos.NacosGetInstances(serviceName)      // 请求中获取实例
	} else { // 如果没有到达则自增
		nacos.RequestTmzMap.Store(serviceName, load.(int)+1)
	}
}

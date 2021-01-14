/**
 * @Time : 2021/1/13 6:20 下午
 * @Author : MassAdobe
 * @Description: filter
**/
package filter

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"github.com/MassAdobe/go-gateway/utils"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 5:40 下午
 * @Description: 黑白名单
**/
func BlackWhiteList(realIp string) {
	if len(realIp) == 0 {
		logs.Lg.Error("黑白名单", errors.New("real ip is nil error"), logs.Desc("当前请求的客户端真实IP为空"))
		panic(errs.NewError(errs.ErrRealIpCode))
	}
	logs.Lg.Debug("黑白名单", logs.Desc(fmt.Sprintf("开始校验，IP为: %s", realIp)))
	// 校验IP地址是否正确
	if utils.CheckIp(realIp) {
		// 白名单优先级高
		if _, okay := nacos.WhiteList[realIp]; okay {
			logs.Lg.Debug("黑白名单", logs.Desc(fmt.Sprintf("命中白名单，IP为: %s", realIp)))
			return
		}
		// 黑名单优先级弱
		if _, okay := nacos.BlackList[realIp]; okay {
			logs.Lg.Debug("黑白名单", logs.Desc(fmt.Sprintf("命中黑名单，禁止后续请求，IP为: %s", realIp)))
			panic(errs.NewError(errs.ErrBlackListCode))
		}
		return
	}
	logs.Lg.Debug("黑白名单", logs.Desc(fmt.Sprintf("当前请求，real-ip错误，IP为: %s", realIp)))
	panic(errs.NewError(errs.ErrRealIpCode))
}

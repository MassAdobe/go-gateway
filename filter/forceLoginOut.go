/**
 * @Time : 2021/1/18 11:29 上午
 * @Author : MassAdobe
 * @Description: filter
**/
package filter

import (
	"fmt"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"github.com/MassAdobe/go-gateway/pojo"
	"github.com/MassAdobe/go-gateway/utils"
	"time"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/18 11:30 上午
 * @Description: 执行强制下线 TODO
**/
func ForceLoginOut(user *pojo.RequestUser, lgTm string) {
	logs.Lg.Debug("执行强制下线", logs.Desc("进入执行强制下线"))
	if user != nil && len(lgTm) != 0 && len(nacos.ForceLoginOut) != 0 { // 命中全局下线
		if tm, okay := nacos.ForceLoginOut[-1]; okay { // 如果存在全局
			logs.Lg.Debug("执行强制下线", logs.Desc(fmt.Sprintf("命中全局下线, 用户: %d, 登录时间: %s", user.UserId, lgTm)))
			if st, err := time.Parse(utils.TimeFormatMS, lgTm); err != nil { // string 转 time
				logs.Lg.Error("执行强制下线", err, logs.Desc(fmt.Sprintf("用户的登录时间解析错误: 用户ID: %d, 用户登录时间: %s", user.UserId, lgTm)))
				panic(errs.NewError(errs.ErrUserLoginTmCode))
			} else {               // 检验时间
				if st.Before(tm) { // 如果当前登录时间在设置超时时间之前
					logs.Lg.Debug("执行强制下线", logs.Desc(fmt.Sprintf("当前用户: %d的登录时间: %s在强制下线时间: %v之前, 需要被强制下线", user.UserId, lgTm, tm)))
					panic(errs.NewError(errs.ErrUserForceLoginOutCode))
				}
			}
		} else {
			if tmm, okay := nacos.ForceLoginOut[user.UserId]; okay {
				logs.Lg.Debug("执行强制下线", logs.Desc(fmt.Sprintf("命中局部下线, 用户: %d, 登录时间: %s", user.UserId, lgTm)))
				if st, err := time.Parse(utils.TimeFormatMS, lgTm); err != nil { // string 转 time
					logs.Lg.Error("执行强制下线", err, logs.Desc(fmt.Sprintf("用户的登录时间解析错误: 用户ID: %d, 用户登录时间: %s", user.UserId, lgTm)))
					panic(errs.NewError(errs.ErrUserLoginTmCode))
				} else {                // 检验时间
					if st.Before(tmm) { // 如果当前登录时间在设置超时时间之前
						logs.Lg.Debug("执行强制下线", logs.Desc(fmt.Sprintf("当前用户: %d的登录时间: %s在强制下线时间: %v之前, 需要被强制下线", user.UserId, lgTm, tmm)))
						delete(nacos.ForceLoginOut, user.UserId) // 删除当前用户的下线
						panic(errs.NewError(errs.ErrUserForceLoginOutCode))
					}
				}
			}
		}
	}
}

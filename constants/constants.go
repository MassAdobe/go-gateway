/**
 * @Time : 2020-04-26 19:57
 * @Author : MassAdobe
 * @Description: config
**/
package constants

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 2:02 下午
 * @Description: HTTP中的基本常量
**/
const (
	CONTENT_TYPE_KEY   = "Content-Type"                   // 请求协议种类键值
	CONTENT_TYPE_INNER = "application/json;charset=utf-8" // 请求协议种类内容
	REQUEST_USER_KEY   = "user"                           // 用户头信息键值
	REQUEST_REAL_HOST  = "Real-Host"                      // 真实服务请求地址键值
	REQUEST_REAL_IP    = "X-Real-Ip"                      // 真实请求IP地址键值
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 1:56 下午
 * @Description: jwt常量
**/
const (
	TOKEN_KEY          = "access-token" // header中access-token名称
	TOKEN_USER_KEY     = "usr_id"       // Token中的用户KEY
	TOKEN_LOGIN_TM_KEY = "lgn_tm"       // Token中的Login时间
	TOKEN_USER_FROM    = "usr_frm"      // 用户登录来源
)

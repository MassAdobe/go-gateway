/**
 * @Time : 2020-04-26 19:57
 * @Author : MassAdobe
 * @Description: config
**/
package errs

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:05
 * @Description: 错误封装
**/
const (
	/*error code*/
	ErrSystemCode            = 900 + iota // 内部错误
	ErrServiceNilCode                     // 当前不存在该服务
	ErrGatewayCode                        // 网关错误
	ErrConnectRefusedCode                 // 调用服务失败
	ErrNacosGetInstanceCode               // 获取注册中心服务失败
	ErrBlackListCode                      // 命中黑名单，非法请求
	ErrRealIpCode                         // 当前客户端请求IP错误，非法请求
	ErrTokenCode                          // 当前请求携带TOKEN错误
	ErrTokenLoginTmCode                   // 登录过期
	ErrJsonCode                           // json解析错误
	ErrUserLoginTmCode                    // 用户登录时间异常
	ErrUserForceLoginOutCode              // 强制下线

	/*error desc*/
	ErrSystemDesc            = "内部错误"
	ErrServiceNilDesc        = "当前不存在该服务"
	ErrGatewayDesc           = "网关错误"
	ErrConnectRefusedDesc    = "调用服务失败"
	ErrNacosGetInstanceDesc  = "获取注册中心服务失败"
	ErrBlackListDesc         = "命中黑名单，非法请求"
	ErrRealIpDesc            = "当前客户端请求IP错误，非法请求"
	ErrTokenDesc             = "当前请求携带TOKEN错误"
	ErrTokenLoginTmDesc      = "登录过期"
	ErrJsonDesc              = "Json转换错误"
	ErrUserLoginTmDesc       = "用户登录时间异常"
	ErrUserForceLoginOutDesc = "强制下线"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:06
 * @Description: 错误参数体
**/
var CodeDescMap = map[int]string{
	// 系统错误
	ErrSystemCode:            ErrSystemDesc,
	ErrServiceNilCode:        ErrServiceNilDesc,
	ErrGatewayCode:           ErrGatewayDesc,
	ErrConnectRefusedCode:    ErrConnectRefusedDesc,
	ErrNacosGetInstanceCode:  ErrNacosGetInstanceDesc,
	ErrBlackListCode:         ErrBlackListDesc,
	ErrRealIpCode:            ErrRealIpDesc,
	ErrTokenCode:             ErrTokenDesc,
	ErrTokenLoginTmCode:      ErrTokenLoginTmDesc,
	ErrJsonCode:              ErrJsonDesc,
	ErrUserLoginTmCode:       ErrUserLoginTmDesc,
	ErrUserForceLoginOutCode: ErrUserForceLoginOutDesc,
}

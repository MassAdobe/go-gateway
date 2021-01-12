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
	SuccessCode           = 0    // 成功
	ErrSystemCode         = iota // 内部错误
	ErrServiceNilCode            // 当前不存在该服务
	ErrTestCode                  // 测试错误
	ErrGatewayCode               // 网关错误
	ErrConnectRefusedCode        // 调用服务失败

	/*error desc*/
	SuccessDesc           = "成功"
	ErrSystemDesc         = "内部错误"
	ErrServiceNilDesc     = "当前不存在该服务"
	ErrTestDesc           = "测试错误"
	ErrGatewayDesc        = "网关错误"
	ErrConnectRefusedDesc = "调用服务失败"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:06
 * @Description: 错误参数体
**/
var CodeDescMap = map[int]string{
	// 系统错误
	SuccessCode:           SuccessDesc,
	ErrSystemCode:         ErrSystemDesc,
	ErrServiceNilCode:     ErrServiceNilDesc,
	ErrTestCode:           ErrTestDesc,
	ErrGatewayCode:        ErrGatewayDesc,
	ErrConnectRefusedCode: ErrConnectRefusedDesc,
}

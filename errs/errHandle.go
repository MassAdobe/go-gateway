/**
 * @Time : 2021/1/12 11:38 上午
 * @Author : MassAdobe
 * @Description: errs
**/
package errs

import (
	"net/http"
)

var (
	ServerError = BasicNewError(http.StatusInternalServerError, ErrGatewayCode, "", nil)
)

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-27 20:23
 * @Description: 创建新异常
**/
func NewError(code int, errs ...error) *Error {
	if len(errs) != 0 {
		return BasicNewError(http.StatusOK, code, "", errs[0])
	}
	return BasicNewError(http.StatusOK, code, "", nil)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-27 20:20
 * @Description: 基类异常
**/
func BasicNewError(status, code int, msg string, err error) *Error {
	if len(msg) == 0 {
		return &Error{
			StatusCode: status,
			Code:       code,
			Msg:        CodeDescMap[code],
			Data:       "",
		}
	}
	return &Error{
		StatusCode: status,
		Code:       code,
		Msg:        CodeDescMap[code] + "\n" + msg,
		Data:       "",
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-27 20:14
 * @Description: 错误处理的结构体
**/
type Error struct {
	StatusCode int         `json:"status"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-27 20:25
 * @Description: 其他错误
**/
func OtherError(message string) *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		Code:       ErrGatewayCode,
		Msg:        CodeDescMap[ErrGatewayCode],
		Data:       message,
	}
}

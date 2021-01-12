/**
 * @Time : 2021/1/12 9:26 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"net/http/httputil"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 9:36 上午
 * @Description: 反向代理方法
**/
func NewMultipleHostsReverseProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director:     rtnDirector(), // 请求协调者
		ErrorHandler: rtnFailure(),  // 错误回调
	}
}

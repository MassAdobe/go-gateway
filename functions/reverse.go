/**
 * @Time : 2021/1/12 9:26 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/logs"
	"net/http"
	"net/http/httputil"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 9:36 上午
 * @Description: 反向代理方法
**/
func NewMultipleHostsReverseProxy() http.HandlerFunc {
	var rp = &httputil.ReverseProxy{
		Director:     rtnDirector(), // 请求协调者
		Transport:    transport,     // 反向代理配置
		ErrorHandler: rtnFailure(),  // 错误回调
	}
	return func(wr http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var Err *errs.Error
				if e, ok := err.(*errs.Error); ok {
					Err = e
				} else if e, ok := err.(error); ok {
					Err = errs.OtherError(e.Error())
				} else {
					Err = errs.ServerError
				}
				logs.Lg.Error("全局错误", errors.New("global error occurred"), logs.Desc(fmt.Sprintf("错误为: %v", err)))
				wr.Header().Set(constants.CONTENT_TYPE_KEY, constants.CONTENT_TYPE_INNER)
				marshal, _ := json.Marshal(Err)
				_, _ = wr.Write(marshal)
			}
		}()
		rp.ServeHTTP(wr, req)
	}
}

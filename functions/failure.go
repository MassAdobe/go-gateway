/**
 * @Time : 2021/1/12 11:20 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"encoding/json"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"net/http"
	"net/url"
	"strings"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 11:21 上午
 * @Description: 返回错误回调
**/
func rtnFailure() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(write http.ResponseWriter, req *http.Request, err error) {
		logs.Lg.Error("返回错误回调", err)
		// 如果是连接断掉，那么需要清理不能用的连接
		if strings.Contains(err.Error(), "connection refused") {
			index := strings.Index(req.RequestURI[1:], "/")
			serviceName := req.RequestURI[1 : index+1]
			load, _ := nacos.Instances.Load(serviceName)
			newUrls := make([]*url.URL, 0)
			host := req.Header.Get(constants.REQUEST_REAL_HOST)
			for _, v := range load.([]*url.URL) {
				if v.Host != host {
					newUrls = append(newUrls, &url.URL{
						Scheme: nacos.DEFAULT_SCHEMA,
						Host:   v.Host,
					})
				}
			}
			nacos.Instances.Store(serviceName, newUrls)
			write.Header().Set(constants.CONTENT_TYPE_KEY, constants.CONTENT_TYPE_INNER)
			marshal, _ := json.Marshal(errs.NewError(errs.ErrConnectRefusedCode))
			write.Write(marshal)
		}
	}
}

/**
 * @Time : 2021/1/12 11:20 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"encoding/json"
	"fmt"
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
		index := strings.Index(req.RequestURI[1:], constants.BACKSLASH_MARK)
		serviceName := req.RequestURI[1 : index+1]
		write.Header().Set(constants.CONTENT_TYPE_KEY, constants.CONTENT_TYPE_INNER)
		var rtn []byte
		// 如果是连接断掉，那么需要清理不能用的连接
		if strings.Contains(err.Error(), "connection refused") {
			logs.Lg.Error("返回错误回调", err, logs.Desc(fmt.Sprintf("当前错误是: connection refused; 服务: %s, 请求: %s, 调用方: %s", serviceName, req.RequestURI, req.Host)))
			load, _ := nacos.Instances.Load(serviceName)
			newUrls := make([]*url.URL, 0)
			host := req.Header.Get(constants.REQUEST_REAL_HOST)
			for _, v := range load.([]*url.URL) {
				if v.Host != host {
					newUrls = append(newUrls, &url.URL{
						Scheme: constants.NACOS_SCHEMA,
						Host:   v.Host,
					})
				}
			}
			nacos.Instances.Store(serviceName, newUrls)
			logs.Lg.Debug("返回错误回调", logs.Desc("删除相关服务调用HOST"))
			rtn, _ = json.Marshal(errs.NewError(errs.ErrConnectRefusedCode, err))
		} else { // 其他报错
			logs.Lg.Error("返回错误回调", err, logs.Desc(fmt.Sprintf("服务: %s, 请求: %s, 调用方: %s", serviceName, req.RequestURI, req.Host)))
			rtn, _ = json.Marshal(errs.NewError(errs.ErrGatewayCode, err))
		}
		_, _ = write.Write(rtn)
	}
}

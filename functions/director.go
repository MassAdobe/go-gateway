/**
 * @Time : 2021/1/12 11:20 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/filter"
	"github.com/MassAdobe/go-gateway/loadbalance"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"net/http"
	"net/url"
	"strings"
)

const (
	DEFAULT_SCHEMA = "http" // 默认转发方式
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 10:53 上午
 * @Description: 返回请求协调者
**/
func rtnDirector() func(req *http.Request) {
	return func(req *http.Request) {
		logs.Lg.Debug("请求协调者", logs.Desc(fmt.Sprintf("请求来自于: %s, 请求资源: %s", req.RemoteAddr, req.RequestURI)))
		filter.BlackWhiteList(req) // 黑白名单
		index := strings.Index(req.RequestURI[1:], "/")
		serviceName := req.RequestURI[1 : index+1]
		{
			// TODO 整理头信息(暂时没有确定相关登录方法，暂时不写)
		}
		if nacos.PuGrayScale.Open { // 如果开启灰度
			// TODO 一些校验条件
			if loadbalance.Lb.Type == "nacos" { // 基于nacos的WRR负载 灰度
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用灰度发布下nacos负载"))
				grayScaleNacosDirector(req, serviceName)
			} else { // 基于自研的负载 灰度
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用灰度发布下自研负载"))
				// TODO 自研灰度
			}
		} else {                                // 关闭灰度
			if loadbalance.Lb.Type == "nacos" { // 基于nacos的WRR负载
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用nacos负载"))
				nacosDirector(req, serviceName)
			} else { // 基于自研的负载
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用自研负载"))
				selfDirector(req, serviceName)
			}
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 8:54 下午
 * @Description: 灰度发布下的nacos
**/
func grayScaleNacosDirector(req *http.Request, serviceName string) {
	if server, err := nacos.NacosGetServer(serviceName, nacos.InitConfiguration.Routers.Services[serviceName], nacos.PuGrayScale.Version); err != nil {
		logs.Lg.Error("返回请求协调者", err, logs.Desc(fmt.Sprintf("请求的服务名: %s", serviceName)))
		panic(errs.NewError(errs.ErrServiceNilCode))
	} else {
		target := &url.URL{
			Scheme: DEFAULT_SCHEMA,
			Host:   fmt.Sprintf("%s:%d", server.Ip, server.Port),
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Header.Set(constants.REQUEST_REAL_HOST, target.Host)
		req.Header.Set(constants.REQUEST_REAL_IP, req.Header.Get(constants.REQUEST_REAL_IP))
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 10:49 上午
 * @Description: 基于nacos的WRR负载
**/
func nacosDirector(req *http.Request, serviceName string) {
	if server, err := nacos.NacosGetServer(serviceName, nacos.InitConfiguration.Routers.Services[serviceName], ""); err != nil {
		logs.Lg.Error("返回请求协调者", err, logs.Desc(fmt.Sprintf("请求的服务名: %s", serviceName)))
		panic(errs.NewError(errs.ErrServiceNilCode))
	} else {
		target := &url.URL{
			Scheme: DEFAULT_SCHEMA,
			Host:   fmt.Sprintf("%s:%d", server.Ip, server.Port),
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Header.Set(constants.REQUEST_REAL_HOST, target.Host)
		req.Header.Set(constants.REQUEST_REAL_IP, req.Header.Get(constants.REQUEST_REAL_IP))
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 10:50 上午
 * @Description: 基于自研的负载
**/
func selfDirector(req *http.Request, serviceName string) {
	var urls interface{}
	var okay bool
	if urls, okay = nacos.Instances.Load(serviceName); !okay || nil == urls { // 如果不存在
		go nacos.NacosGetInstances(serviceName) // 请求中获取实例
	}
	if nil == urls {
		logs.Lg.Error("返回请求协调者", errors.New("urls is nil"))
		panic(errs.NewError(errs.ErrServiceNilCode))
	}
	target := loadbalance.Lb.CurUrl(serviceName, urls) // 根据当前选择，返回url
	if target == nil {
		logs.Lg.Error("返回请求协调者", errors.New("no service is available"), logs.Desc(fmt.Sprintf("请求的服务名: %s", serviceName)))
		panic(errs.NewError(errs.ErrServiceNilCode))
	}
	// 整理请求地址
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Header.Set(constants.REQUEST_REAL_HOST, target.Host)
	req.Header.Set(constants.REQUEST_REAL_IP, req.Header.Get(constants.REQUEST_REAL_IP))
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	go filter.CheckTmz(serviceName) // 检查次数，如果超过了相关次数，重新获取服务信息
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 3:32 下午
 * @Description: 拼接请求地址
**/
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

/**
 * @Time : 2021/1/12 11:20 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/filter"
	"github.com/MassAdobe/go-gateway/loadbalance"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"net/http"
	"net/url"
	"strings"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 10:53 上午
 * @Description: 返回请求协调者
**/
func rtnDirector() func(req *http.Request) {
	return func(req *http.Request) {
		logs.Lg.Debug("请求")
		index := strings.Index(req.RequestURI[1:], "/")
		serviceName := req.RequestURI[1 : index+1]
		if loadbalance.Lb.Type == "nacos" { // 基于nacos的WRR负载
			nacosDirector(req, serviceName)
		} else { // 基于自研的负载
			selfDirector(req, serviceName)
		}
		// TODO 整理头信息(暂时没有确定相关登录方法，暂时不写)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 10:49 上午
 * @Description: 基于nacos的WRR负载
**/
func nacosDirector(req *http.Request, serviceName string) {
	if server, err := nacos.NacosGetServer(serviceName, nacos.InitConfiguration.Routers.Services[serviceName]); err != nil {
		logs.Lg.Error("返回请求协调者", err)
	} else {
		target := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", server.Ip, server.Port),
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Header.Set(constants.REQUEST_REAL_HOST, target.Host)
		req.Header.Set(constants.REQUEST_REAL_IP, req.RemoteAddr)
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
	if urls, okay = nacos.Instances.Load(serviceName); !okay { // 如果不存在
		go nacos.NacosGetInstances(serviceName) // 请求中获取实例
	}
	target := loadbalance.Lb.CurUrl(serviceName, urls) // 根据当前选择，返回url
	// TODO 待解决 如果请求为空 需要返回不能用的问题
	if target == nil {
		return
	}
	// 整理请求地址
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Header.Set(constants.REQUEST_REAL_HOST, target.Host)
	req.Header.Set(constants.REQUEST_REAL_IP, req.RemoteAddr)
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

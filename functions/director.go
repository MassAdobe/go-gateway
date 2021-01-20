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
	"github.com/MassAdobe/go-gateway/pojo"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 10:53 上午
 * @Description: 返回请求协调者
**/
func rtnDirector() func(req *http.Request) {
	return func(req *http.Request) {
		logs.Lg.Debug("请求协调者", logs.Desc(fmt.Sprintf("请求来自于: %s, 请求资源: %s", req.RemoteAddr, req.RequestURI)))
		realIp := req.Header.Get(constants.REQUEST_REAL_IP) // 获取真实IP
		filter.BlackWhiteList(realIp)                       // 黑白名单
		user, lgTm := filter.VerifiedJWT(req)               // 校验jwt的token(同时返回用户信息)
		filter.ForceLoginOut(user, lgTm)                    // 强制下线
		index := strings.Index(req.RequestURI[1:], constants.BACKSLASH_MARK)
		serviceName := req.RequestURI[1 : index+1]
		// 灰度开启情况
		if nacos.PuGrayScale.Open { // 如果开启灰度
			switch nacos.PuGrayScale.Type {
			case constants.GRAY_SCALE_IP_LIST_TYPE: // IP列表
				// 存在于灰度发布的列表中 直接走灰度的路由
				if _, okay := nacos.PuGrayScale.List[realIp]; okay {
					break
				}
				goto Loop // 不存在于灰度发布的列表中 走正常路由
			case constants.GRAY_SCALE_USER_LIST_TYPE: // 用户列表
				if user != nil { // 当前用户不为空
					userId := strconv.FormatInt(user.UserId, 10)
					// 存在与灰度发布的列表中 直接走灰度的路由
					if _, okay := nacos.PuGrayScale.List[userId]; okay {
						break
					}
				}
				goto Loop // 不存在于灰度发布的列表中 走正常路由
			case constants.GRAY_SCALE_USER_ID_TYPE: // 用户范围
				// 存在与灰度发布的用户范围中 直接走灰度的路由
				if userScopeCheck(user) {
					break
				}
				goto Loop // 不存在于灰度发布的列表中 走正常路由
			default: // 默认不走灰度
				goto Loop // 不存在于灰度发布的列表中 走正常路由
			}
			if strings.ToLower(loadbalance.Lb.Type) == constants.NACOS_MARK { // 基于nacos的WRR负载 灰度
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用灰度发布下nacos负载"))
				grayScaleNacosDirector(req, serviceName)
			} else { // 基于自研的负载 灰度
				logs.Lg.Debug("请求协调者", logs.Desc("当前请求使用灰度发布下自研负载"))
				grayScaleSelfDirector(req, serviceName)
			}
			return
		}
		// 非灰度开启情况
	Loop:
		{
			if loadbalance.Lb.Type == constants.NACOS_MARK { // 基于nacos的WRR负载
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
 * @TIME: 2021/1/18 10:22 上午
 * @Description: 校验用户scope的灰度
**/
func userScopeCheck(user *pojo.RequestUser) bool {
	if user != nil {
		switch strings.ToLower(nacos.PuGrayScale.Scope.Type) {
		case constants.GRAY_SCALE_USER_SCOPE_GREAT: // 用户范围灰度：大于
			if user.UserId >= nacos.PuGrayScale.Scope.Mark {
				return true
			}
			break
		case constants.GRAY_SCALE_USER_SCOPE_LESS: // 用户范围灰度：小于
			if user.UserId <= nacos.PuGrayScale.Scope.Mark {
				return true
			}
			break
		default: // 默认不走灰度
			break
		}
	}
	return false
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
			Scheme: constants.DEFAULT_SCHEMA,
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
			req.URL.RawQuery = targetQuery + constants.AND_MARK + req.URL.RawQuery
		}
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 10:50 上午
 * @Description: 基于自研的负载
**/
func grayScaleSelfDirector(req *http.Request, serviceName string) {
	var urls interface{}
	var okay bool
	if urls, okay = nacos.GrayScaleInstances.Load(serviceName); !okay || nil == urls { // 如果不存在
		go nacos.NacosGetGrayScaleInstances(serviceName) // 请求中获取实例
	}
	if nil == urls {
		logs.Lg.Error("返回请求协调者", errors.New("urls is nil"))
		panic(errs.NewError(errs.ErrServiceNilCode))
	}
	target := loadbalance.Lb.GrayScaleCurUrl(serviceName, urls) // 根据当前选择，返回url
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
		req.URL.RawQuery = targetQuery + constants.AND_MARK + req.URL.RawQuery
	}
	go filter.CheckGrayScaleTmz(serviceName) // 灰度，检查次数，如果超过了相关次数，重新获取服务信息
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 10:49 上午
 * @Description: 基于nacos的WRR负载
**/
func nacosDirector(req *http.Request, serviceName string) {
	if server, err := nacos.NacosGetServer(serviceName, nacos.InitConfiguration.Routers.Services[serviceName], nacos.Version); err != nil {
		logs.Lg.Error("返回请求协调者", err, logs.Desc(fmt.Sprintf("请求的服务名: %s", serviceName)))
		panic(errs.NewError(errs.ErrServiceNilCode))
	} else {
		target := &url.URL{
			Scheme: constants.DEFAULT_SCHEMA,
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
			req.URL.RawQuery = targetQuery + constants.AND_MARK + req.URL.RawQuery
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
		req.URL.RawQuery = targetQuery + constants.AND_MARK + req.URL.RawQuery
	}
	go filter.CheckTmz(serviceName) // 检查次数，如果超过了相关次数，重新获取服务信息
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 3:32 下午
 * @Description: 拼接请求地址
**/
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, constants.BACKSLASH_MARK)
	bslash := strings.HasPrefix(b, constants.BACKSLASH_MARK)
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

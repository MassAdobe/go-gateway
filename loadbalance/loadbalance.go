/**
 * @Time : 2021/1/12 2:49 下午
 * @Author : MassAdobe
 * @Description: functions
**/
package loadbalance

import (
	"errors"
	"github.com/MassAdobe/go-gateway/logs"
	"math/rand"
	"net/url"
	"sync"
)

const (
	LOAD_BALANCE_ROUND  = "round"  // 轮询
	LOAD_BALANCE_RANDOM = "random" // 随机
)

var (
	Lb *LoadBalance
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 2:50 下午
 * @Description: 负载均衡实体
**/
type LoadBalance struct {
	Type  string   // 种类："random":随机;"round":轮询
	Round sync.Map // 如果是轮询，那么记录轮询值
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 3:04 下午
 * @Description: 随机
**/
func (this *LoadBalance) RandomLB(urls interface{}) *url.URL {
	if len(urls.([]*url.URL)) == 0 {
		logs.Lg.Error("负载均衡", errors.New("current request has not url target"))
		return nil
	}
	return urls.([]*url.URL)[rand.Int()%len(urls.([]*url.URL))]
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 3:04 下午
 * @Description: 轮询
**/
func (this *LoadBalance) RoundLB(serviceName string, urls interface{}) *url.URL {
	if curInt, okay := this.Round.Load(serviceName); okay {
		urlList := urls.([]*url.URL)
		if len(urls.([]*url.URL))-1 < curInt.(int) {
			return urlList[0]
		} else if len(urls.([]*url.URL))-1 == curInt.(int) {
			this.Round.Store(serviceName, 0)
		} else {
			this.Round.Store(serviceName, curInt.(int)+1)
		}
		return urlList[curInt.(int)]
	}
	logs.Lg.Error("负载均衡", errors.New("current request has not url target"))
	return nil
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 3:08 下午
 * @Description: 根据当前选择，返回url
**/
func (this *LoadBalance) CurUrl(serviceName string, urls interface{}) *url.URL {
	if this.Type == LOAD_BALANCE_RANDOM { // 随机
		return this.RandomLB(urls)
	} else if this.Type == LOAD_BALANCE_ROUND { // 轮询
		return this.RoundLB(serviceName, urls)
	} else { // 都没有
		logs.Lg.Error("负载均衡", errors.New("have not select load balance type"))
		return nil
	}
}

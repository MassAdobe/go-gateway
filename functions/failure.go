/**
 * @Time : 2021/1/12 11:20 上午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"net/http"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/12 11:21 上午
 * @Description: 返回错误回调
**/
func rtnFailure() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
	}
}

/**
 * @Time : 2021/1/13 3:31 下午
 * @Author : MassAdobe
 * @Description: functions
**/
package functions

import (
	"net/http"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 3:32 下午
 * @Description: 更改返回内容
**/
func rtnResponse() func(rsp *http.Response) error {
	return func(rsp *http.Response) error {
		//if rsp.StatusCode != 200 {
		//	//获取内容
		//	oldPayload, err := ioutil.ReadAll(rsp.Body)
		//	if err != nil {
		//		return err
		//	}
		//	//追加内容
		//	newPayload := []byte("StatusCode error:" + string(oldPayload))
		//	rsp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
		//	rsp.ContentLength = int64(len(newPayload))
		//	rsp.Header.Set("Content-Length", strconv.FormatInt(int64(len(newPayload)), 10))
		//}
		return nil
	}
}
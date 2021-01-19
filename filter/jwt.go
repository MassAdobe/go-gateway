/**
 * @Time : 2021/1/15 1:59 下午
 * @Author : MassAdobe
 * @Description: filter
**/
package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/errs"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"github.com/MassAdobe/go-gateway/pojo"
	"github.com/MassAdobe/go-gateway/utils"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

const (
	FULL_STOP_MARK = "."
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 1:59 下午
 * @Description: 校验jwt的token
**/
func VerifiedJWT(req *http.Request) (*pojo.RequestUser, string) {
	token, lgTm := decodeToken(req.Header.Get(constants.TOKEN_KEY)), ""
	if len(token) != 0 {
		if tkn, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				errM := errors.New("signing method error")
				logs.Lg.Error("jwt-token", errM, logs.Desc("校验token加密方法错误"))
				return nil, errM
			}
			return []byte(nacos.InitConfiguration.AccessToken.Verify), nil
		}); err != nil {
			panic(errs.NewError(errs.ErrTokenCode))
		} else {
			if tkn.Valid {
				claims, requestUser := tkn.Claims.(jwt.MapClaims), new(pojo.RequestUser)
				// user_id
				if userId, okay := claims[constants.TOKEN_USER_KEY]; okay {
					requestUser.UserId = int64(userId.(float64))
					logs.Lg.Debug("jwt-token", logs.Desc(fmt.Sprintf("获取user_id: %d", requestUser.UserId)))
				} else {
					logs.Lg.Error("jwt-token", errors.New("token has not user_id"), logs.Desc("token中不含有用户ID信息"))
					panic(errs.NewError(errs.ErrTokenCode))
				}
				// user_from
				if userFrom, okay := claims[constants.TOKEN_USER_FROM]; okay {
					requestUser.UserFrom = userFrom.(string)
					logs.Lg.Debug("jwt-token", logs.Desc(fmt.Sprintf("user_from: %s", requestUser.UserFrom)))
				} else {
					logs.Lg.Error("jwt-token", errors.New("token has not user_from"), logs.Desc("token中不含有用户来源信息"))
					panic(errs.NewError(errs.ErrTokenCode))
				}
				// login_tm
				if loginTm, okay := claims[constants.TOKEN_LOGIN_TM_KEY]; okay {
					logs.Lg.Debug("jwt-token", logs.Desc(fmt.Sprintf("login_tm: %s", loginTm.(string))))
					lgTm = loginTm.(string)
					differ := utils.GetHourDiffer(lgTm)
					// 大于token更新时间，小于过期时间，更新token
					if nacos.InitConfiguration.AccessToken.Refresh <= differ && differ <= nacos.InitConfiguration.AccessToken.Expire {
						logs.Lg.Debug("jwt-token", logs.Desc("当前token时间可以更新token"))
						req.Header.Set(constants.TOKEN_KEY, createToken(requestUser.UserId, requestUser.UserFrom))
					} else if nacos.InitConfiguration.AccessToken.Expire <= differ { // 大于过期时间，直接过期
						logs.Lg.Error("jwt-token", errors.New("token login time is expired"), logs.Desc("当前token时间已经过期"))
						panic(errs.NewError(errs.ErrTokenLoginTmCode))
					} else { // 正常度过
						req.Header.Del(constants.TOKEN_KEY) // 正常就删除头中的token信息
						logs.Lg.Debug("jwt-token", logs.Desc("当前token时间正常"))
					}
				} else {
					logs.Lg.Error("jwt-token", errors.New("token has not login_tm"), logs.Desc("token中不含有用户登录时间信息"))
					panic(errs.NewError(errs.ErrTokenCode))
				}
				// 放入用户信息
				if user, err := json.Marshal(requestUser); err != nil {
					logs.Lg.Error("jwt-token", err)
					panic(errs.NewError(errs.ErrJsonCode))
				} else {
					req.Header.Set(constants.REQUEST_USER_KEY, string(user))
				}
				return requestUser, lgTm
			}
			logs.Lg.Error("jwt-token", errors.New("token verified error"), logs.Desc("校验token错误"))
		}
	}
	logs.Lg.Debug("jwt-token", logs.Desc(fmt.Sprintf("请求: %s, 来源: %s, 不存在access-token", req.RequestURI, req.Header.Get(constants.REQUEST_REAL_IP))))
	return nil, ""
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 1:54 下午
 * @Description: 生成Access-Token
**/
func createToken(userId int64, userFrom string) string {
	claim := jwt.MapClaims{
		constants.TOKEN_USER_KEY:     userId,
		constants.TOKEN_USER_FROM:    userFrom,
		constants.TOKEN_LOGIN_TM_KEY: utils.RtnTmString(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	if tokenss, err := token.SignedString([]byte(nacos.InitConfiguration.AccessToken.Verify)); err != nil {
		logs.Lg.Error("jwt-token", err, logs.Desc("生成access-token错误"))
	} else {
		return encodeToken(tokenss)
	}
	return ""
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 2:58 下午
 * @Description: 加密token
**/
func encodeToken(token string) string {
	split, rtn := strings.Split(token, FULL_STOP_MARK), ""
	for idx, str := range split {
		if 0 != idx {
			split[idx] = str[10:] + str[:10]
			rtn += FULL_STOP_MARK + split[idx]
		} else {
			rtn += split[idx]
		}
	}
	return rtn
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 2:58 下午
 * @Description: 解密token
**/
func decodeToken(token string) string {
	split, rtn := strings.Split(token, FULL_STOP_MARK), ""
	for idx, str := range split {
		if 0 != idx {
			split[idx] = str[len(str)-10:] + str[:len(str)-10]
			rtn += FULL_STOP_MARK + split[idx]
		} else {
			rtn += split[idx]
		}
	}
	return rtn
}

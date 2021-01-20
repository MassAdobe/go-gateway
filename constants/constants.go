/**
 * @Time : 2020-04-26 19:57
 * @Author : MassAdobe
 * @Description: config
**/
package constants

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/20 10:28 上午
 * @Description: 主进程常量
**/
const (
	SYSTEM_CONTROL_PWD = "pwd"
	CONFIG_NAME        = "/config.yml"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 2:02 下午
 * @Description: HTTP中的基本常量
**/
const (
	CONTENT_TYPE_KEY   = "Content-Type"                   // 请求协议种类键值
	CONTENT_TYPE_INNER = "application/json;charset=utf-8" // 请求协议种类内容
	REQUEST_USER_KEY   = "user"                           // 用户头信息键值
	REQUEST_REAL_HOST  = "Real-Host"                      // 真实服务请求地址键值
	REQUEST_REAL_IP    = "X-Real-Ip"                      // 真实请求IP地址键值
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 1:56 下午
 * @Description: jwt常量
**/
const (
	TOKEN_KEY          = "access-token" // header中access-token名称
	TOKEN_USER_KEY     = "usr_id"       // Token中的用户KEY
	TOKEN_LOGIN_TM_KEY = "lgn_tm"       // Token中的Login时间
	TOKEN_USER_FROM    = "usr_frm"      // 用户登录来源
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/20 9:57 上午
 * @Description: 标点符号
**/
const (
	FULL_STOP_MARK = "."
	BACKSLASH_MARK = "/"
	AND_MARK       = "&"
	SPACE_MARK     = " "
	COLON_MARK     = ":"
	QUESTION_MARK  = "?"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/20 10:00 上午
 * @Description: function中常量
**/
const (
	DEFAULT_SCHEMA              = "http"  // 默认转发方式
	GRAY_SCALE_USER_SCOPE_GREAT = "great" // 用户范围灰度：大于
	GRAY_SCALE_USER_SCOPE_LESS  = "less"  // 用户范围灰度：小于
	NACOS_MARK                  = "nacos"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/20 10:02 上午
 * @Description: 负载均衡常量
**/
const (
	LOAD_BALANCE_ROUND  = "round"  // 轮询
	LOAD_BALANCE_RANDOM = "random" // 随机
)

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 20:02
 * @Description: 日志常量
**/
const (
	TIME             = "time"
	LOG_LEVEL        = "log_level"
	LOGGER           = "logger"
	DESC             = "desc"
	MSG              = "msg"
	TRACE            = "trace"
	ERROR            = "error"
	TIME_FORMAT      = "2006-01-02 15:04:05.000"
	SERVER_NAME_MARK = "server_name"
	LOG_LEVEL_DEBUG  = "debug"
	LOG_LEVEL_INFO   = "info"
	LOG_LEVEL_ERROR  = "error"
	FUNCTION_MARK    = "function"
	PATH_NUM_MARK    = "path_num"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/20 10:06 上午
 * @Description: nacos常量
**/
const (
	COMMA_MARK                    = ","
	NACOS_CONTEXT_PATH            = "/nacos"
	NACOS_LOG_DIR                 = "/tmp/nacos/log"
	NACOS_LOG_CACHE_DIR           = "/tmp/nacos/cache"
	NACOS_ROTATE_TIME             = "1h"
	NACOS_MAX_AGE                 = 3
	NACOS_LOG_LEVEL               = "debug"
	NACOS_SCHEMA                  = "http"
	NACOS_NOT_LOAD_CACHE_AT_START = true
	LOG_LEVEL_MODIFIED_DEBUG      = "debug"
	LOG_LEVEL_MODIFIED_INFO       = "info"
	LOG_LEVEL_MODIFIED_WARN       = "warn"
	LOG_LEVEL_MODIFIED_ERROR      = "error"
	LOG_LEVEL_MODIFIED_DPANIC     = "dpanic"
	LOG_LEVEL_MODIFIED_PANIC      = "panic"
	LOG_LEVEL_MODIFIED_FATAL      = "fatal"
	INSTANCE_LIST_EMPTY           = "instance list is empty!" // 列表为空错误
	NACOS_SERVER_CONFIGS_MARK     = "serverConfigs"
	NACOS_CLIENT_CONFIG_MARK      = "clientConfig"
	NACOS_REGIST_IDC_MARK         = "idc"
	NACOS_REGIST_IDC_INNER        = "shanghai"
	NACOS_REGIST_TIMESTAMP_MARK   = "timestamp"
	NACOS_DISCOVERY_CLUSTER_NAME  = "DEFAULT"
	NACOS_CONFIGURATION_MARK      = "nacos"
	GRAY_SCALE_USER_ID_TYPE       = "userscope" // 灰度发布种类：用户ID范围
	GRAY_SCALE_USER_LIST_TYPE     = "userlist"  // 灰度发布种类：用户列表
	GRAY_SCALE_IP_LIST_TYPE       = "iplist"    // 灰度发布种类：IP列表
)

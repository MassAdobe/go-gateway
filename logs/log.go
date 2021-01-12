/**
 * @Time : 2020-04-26 19:57
 * @Author : MassAdobe
 * @Description: logs
**/
package logs

import (
	_ "encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
	"time"
)

var (
	Lg MyLog
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 6:17 下午
 * @Description: 日志对象
**/
type MyLog struct {
	ZapLog *zap.Logger
	Level  zap.AtomicLevel
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 20:02
 * @Description: 日志常量
**/
const (
	TIME        = "time"
	LOG_LEVEL   = "log_level"
	LOGGER      = "logger"
	DESC        = "desc"
	MSG         = "msg"
	TRACE       = "trace"
	ERROR       = "error"
	TIME_FORMAT = "2006-01-02 15:04:05.000"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:03
 * @Description: 新建日志
**/
func NewLogger(filePath, level string, maxSize, maxBackups, maxAge int, compress bool, serviceName string) {
	core := newCore(filePath, level, maxSize, maxBackups, maxAge, compress)
	Lg.ZapLog = zap.New(core, zap.AddCaller(), zap.Development(), zap.Fields(zap.String("server_name", serviceName)))
	zap.ReplaceGlobals(Lg.ZapLog)
	Lg.Info("日志启动成功")
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:04
 * @Description: 日志内部配置
**/
func newCore(filePath, level string, maxSize, maxBackups, maxAge int, compress bool) zapcore.Core {
	// 日志文件路径配置
	hook := lumberjack.Logger{
		Filename:   filePath,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxBackups, // 日志文件最多保存多少个备份
		MaxAge:     maxAge,     // 文件最多保存多少天
		Compress:   compress,   // 是否压缩
	}
	// 设置日志级别
	Lg.Level = zap.NewAtomicLevel()
	switch level {
	case "debug":
		Lg.Level.SetLevel(zap.DebugLevel)
		break
	case "info":
		Lg.Level.SetLevel(zap.InfoLevel)
		break
	case "error":
		Lg.Level.SetLevel(zap.ErrorLevel)
		break
	default:
		Lg.Level.SetLevel(zap.WarnLevel)
		break
	}
	//公用编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        TIME,
		LevelKey:       LOG_LEVEL,
		NameKey:        LOGGER,
		MessageKey:     MSG,
		StacktraceKey:  TRACE,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 颜色编码器
		EncodeTime:     myTimeEncode,                   // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		Lg.Level,                                                                        // 日志级别
	)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:04
 * @Description: 标准化日志日期
**/
func myTimeEncode(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(TIME_FORMAT))
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-27 20:52
 * @Description: 全局处理错误封装
**/
func BasicError(err interface{}) zap.Field {
	return zap.Any(ERROR, err)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:04
 * @Description: 错误封装
**/
func Error(err error) zap.Field {
	if nil == err {
		return zap.String(ERROR, "")
	}
	return zap.String(ERROR, err.Error())
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:04
 * @Description: 描述封装
**/
func Desc(desc interface{}) zap.Field {
	return zap.Any(DESC, desc)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:05
 * @Description: 特殊日志封装
**/
func SpecDesc(name string, desc interface{}) zap.Field {
	return zap.Any(fmt.Sprintf("%s", name), desc)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:05
 * @Description: gin自定义日志插件
**/
func InitLogger(path, serveName, level string, port uint64) {
	NewLogger(fmt.Sprintf("%s/%s-%d", path, serveName, port),
		level,
		128,
		10,
		7,
		true,
		serveName)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 7:52 下午
 * @Description: 重写Global Error日志级别输出
**/
func (this *MyLog) GlobalError(msg string, contextAndFields ...interface{}) {
	fields := this.setTraceAndStep(contextAndFields...)
	if ce := this.ZapLog.Check(zapcore.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 7:52 下午
 * @Description: 重写Debug日志级别输出
**/
func (this *MyLog) Debug(msg string, contextAndFields ...interface{}) {
	fields := this.setTraceAndStep(contextAndFields...)
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fields = append(fields, zap.Any("function", f.Name()))
	fields = append(fields, zap.Any("path_num", fmt.Sprintf("%s:%d", file, line)))
	if ce := this.ZapLog.Check(zapcore.DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 7:52 下午
 * @Description: 重写Info日志级别输出
**/
func (this *MyLog) Info(msg string, contextAndFields ...interface{}) {
	fields := this.setTraceAndStep(contextAndFields...)
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fields = append(fields, zap.Any("function", f.Name()))
	fields = append(fields, zap.Any("path_num", fmt.Sprintf("%s:%d", file, line)))
	if ce := this.ZapLog.Check(zapcore.InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 7:52 下午
 * @Description: 重写Error日志级别输出
**/
func (this *MyLog) Error(msg string, err error, contextAndFields ...interface{}) {
	fields := this.setTraceAndStep(contextAndFields...)
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fields = append(fields, Error(err))
	fields = append(fields, zap.Any("function", f.Name()))
	fields = append(fields, zap.Any("path_num", fmt.Sprintf("%s:%d", file, line)))
	if ce := this.ZapLog.Check(zapcore.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 7:52 下午
 * @Description: 统一输出trace和step信息
**/
func (this *MyLog) setTraceAndStep(contextAndFields ...interface{}) []zap.Field {
	newFields := make([]zap.Field, 0)
	for _, contextAndField := range contextAndFields {
		newFields = append(newFields, contextAndField.(zapcore.Field))
	}
	return newFields
}
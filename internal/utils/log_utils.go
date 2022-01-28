package utils

import (
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

var prodLogger *glog.Logger
var devLogger *glog.Logger
var infoLogger *glog.Logger
var errorLogger *glog.Logger
var handlerLogger *glog.Logger
var asyncDebugLogger *glog.Logger
var writerLogger *glog.Logger

// 使用配置的方式來初始化錯誤日誌
func initErrorLogger() {
	errorLogger = glog.New()
	loggerPath := "/var/log/error"
	err := errorLogger.SetPath(loggerPath)
	if err != nil {
		fmt.Println("init error log path failed, error:" + err.Error())
	}
	errorLogger.SetStdoutPrint(true)
	errorLogger.SetFile("{Y-m-d}[error].log")
	// 打印栈信息
	errorLogger.SetStack(true)
	err = errorLogger.SetLevelStr("ERROR")
	if err != nil {
		fmt.Println("init error log level failed, error:" + err.Error())
	}
	errorLogger.SetWriterColorEnable(true)
}

// 使用Writer接口来收集日志， 我们就打印模拟一下就好了
func initCollectLogger() {
	errorLogger = glog.New()
	loggerPath := "/var/log/critical"
	err := errorLogger.SetPath(loggerPath)
	if err != nil {
		fmt.Println("init critical log path failed, error:" + err.Error())
	}
	errorLogger.SetStdoutPrint(true)
	errorLogger.SetFile("{Y-m-d}[critical].log")
	// 打印栈信息
	errorLogger.SetStack(true)
	err = errorLogger.SetLevelStr("CRIT")
	if err != nil {
		fmt.Println("init writer log level failed, error:" + err.Error())
	}
	errorLogger.SetWriterColorEnable(true)
}

// 链式初始化
func initInfoLogger() {
	infoLogger = glog.New()
	infoLogger.LevelStr("INFO").Path("/var/log/info").Stdout(true).File("{Y-m-d}[info].log").Line(true).SetWriterColorEnable(true)
}

// 自定义handler的日志，可以实现群通知和ES收集
func initHandlerLogger() {
	handlerLogger = glog.New()
	handlerLogger.SetHandlers(loggingJsonHandler)
}

func initAsyncDebugLogger() {
	asyncDebugLogger = glog.New()
	asyncDebugLogger.LevelStr("DEBUG").Path("/var/log/debug").Stdout(true).File("{Y-m-d}[debug].log").Line(true).Async(true).SetWriterColorEnable(true)
	m := make(map[string]interface{}, 3)
	m["RotateExpire"] = 86400000
	err := asyncDebugLogger.SetConfigWithMap(m)
	if err != nil {
		fmt.Println("init debug log level failed, error:" + err.Error())
	}

}

// 选择是否开启debug模式
func EnableDebug(flag bool) {
	if flag == true {
		asyncDebugLogger.SetDebug(true)
	} else {
		asyncDebugLogger.SetDebug(false)
	}
}

func init() {
	initErrorLogger()
	initInfoLogger()
	// 异步日志
	initAsyncDebugLogger()
	// 自定义handler类型的日志
	initHandlerLogger()

	initCollectLogger()

	// 实现writer接口来自定义log操作
	writerLogger.SetWriter(&MyLoggerWriter{logger: writerLogger})

	prodLogger = g.Log("prodLogger")
	devLogger = g.Log("devLogger")
}

// 获取生产日志，采用配置文件的方式
func GetProdLogger() *glog.Logger {
	return prodLogger
}

// 获取开发日志，采用配置文件的方式
func GetDevLogger() *glog.Logger {
	return devLogger
}

// 获取info日志，采用链式初始化的方式
func GetInfoLogger() *glog.Logger {
	return infoLogger
}

// 获取错误日志，采用编码配置的方式
func GetErrorLogger() *glog.Logger {
	return errorLogger
}

// 获取错误日志，采用编码配置的方式
func GetHandlerLogger() *glog.Logger {
	return handlerLogger
}
// 获取自定义日志，采用自定义handler方式
func GetAsyncDebugLogger() *glog.Logger {
	return asyncDebugLogger
}

// 获取自定义日志，采用实现writer接口的方式
func GetWriterLogger() *glog.Logger {
	return writerLogger
}
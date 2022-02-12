package utils

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	PROD     = "prod"
	DEV      = "dev"
	INFO     = "info"
	ERROR    = "error"
	DEBUG    = "debug"
	WARN     = "warn"
	NOTICE   = "notice"
	CRITICAL = "critical"
)

var (
	prodLogger       *glog.Logger
	devLogger        *glog.Logger
	infoLogger       *glog.Logger
	errorLogger      *glog.Logger
	criticalLogger   *glog.Logger
	customizeLogger  *glog.Logger
	asyncDebugLogger *glog.Logger
	warningLogger    *MyLoggerWriter
	logger           *glog.Logger
	myLogger         *MyLoggerWriter
	logPath          *gvar.Var
)

// 使用配置的方式來初始化錯誤日誌
func initErrorLogger() {
	errorLogger = glog.New()
	loggerPath := logPath.String() + ERROR
	err := errorLogger.SetPath(loggerPath)
	if err != nil {
		fmt.Println("init error log path failed, error:" + err.Error())
	}
	errorLogger.SetStdoutPrint(true)
	errorLogger.SetFile("{Y-m-d}[" + ERROR + "].log")
	// 打印栈信息
	errorLogger.SetStack(true)
	err = errorLogger.SetLevelStr("ERROR")
	if err != nil {
		fmt.Println("init error log level failed, error:" + err.Error())
	}
	errorLogger.SetWriterColorEnable(true)
}

// 可以打印栈信息
func initCriticalLogger() {
	criticalLogger = glog.New()
	loggerPath := logPath.String() + CRITICAL
	err := errorLogger.SetPath(loggerPath)
	if err != nil {
		fmt.Println("init critical log path failed, error:" + err.Error())
	}
	errorLogger.SetStdoutPrint(true)
	errorLogger.SetFile("{Y-m-d}[" + CRITICAL + "].log")
	// 打印栈信息
	errorLogger.SetStack(true)
	err = errorLogger.SetLevelStr("CRIT")
	if err != nil {
		fmt.Println("init writer log level failed, error:" + err.Error())
	}
	errorLogger.SetWriterColorEnable(true)
}

// 链式初始化,有大坑
func initInfoLogger() {
	infoLogger = glog.New()
	// 链式初始化是没法改变infoLogger的配置的
	infoLogger.LevelStr("INFO").Path(logPath.String()+INFO).Stdout(true).File("{Y-m-d}["+INFO+"].log").Line(true).Skip(1).Print(nil, "init info logger success.")
	fmt.Println(*infoLogger)
	infoLogger.SetLevel(glog.LEVEL_INFO)
	_ = infoLogger.SetPath(logPath.String() + INFO)
	infoLogger.SetStdoutPrint(true)
	infoLogger.SetFile("{Y-m-d}[" + INFO + "].log")
	infoLogger.SetStackSkip(1)
	infoLogger.SetWriterColorEnable(true)
}

func initCustomizeLog() *glog.Logger {
	w := glog.New()
	m := make(map[string]interface{}, 3)
	m["RotateExpire"] = 86400000
	err := w.SetConfigWithMap(m)
	if err != nil {
		fmt.Println("init notice log level failed, error:" + err.Error())
	}
	w.SetLevel(glog.LEVEL_NOTI)
	w.SetStack(true)
	w.SetWriterColorEnable(true)
	w.SetFile("{Y-m-d}[" + NOTICE + "].log")
	err = w.SetPath(logPath.String() + NOTICE)
	if err != nil {
		fmt.Println("set notice log path failed, error:" + err.Error())
		return nil
	}
	return w
}

// 自定义handler的日志，可以实现群通知和ES收集
func initHandlerLogger() {
	customizeLogger = initCustomizeLog()
	customizeLogger.SetHandlers(loggingJsonHandler)
}

func initAsyncDebugLogger() {
	asyncDebugLogger = glog.New()
	asyncDebugLogger.LevelStr("DEBUG").Path(logPath.String() + DEBUG).Stdout(true).File("{Y-m-d}[" + DEBUG + "].log").Line(true).Async(true).SetWriterColorEnable(true)
	m := make(map[string]interface{}, 3)
	m["RotateExpire"] = 86400000
	err := asyncDebugLogger.SetConfigWithMap(m)
	if err != nil {
		fmt.Println("init debug log level failed, error:" + err.Error())
	}

}

func initWarningLog() *glog.Logger {
	w := glog.New()
	m := make(map[string]interface{}, 3)
	m["RotateExpire"] = 86400000
	err := w.SetConfigWithMap(m)
	if err != nil {
		fmt.Println("init warning log level failed, error:" + err.Error())
	}
	err = w.SetLevelStr("WARN")
	if err != nil {
		fmt.Println("set warning log level failed, error:" + err.Error())
		return nil
	}
	w.SetStack(true)
	w.SetWriterColorEnable(true)
	w.SetFile("{Y-m-d}[" + WARN + "].log")
	err = w.SetPath(logPath.String() + WARN)
	if err != nil {
		fmt.Println("set warning log path failed, error:" + err.Error())
		return nil
	}
	return w
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
	var err error
	logPath, err = g.Cfg("customizeConfig").Get(gctx.New(), "customize.logPath")
	if err != nil {
		panic("get log path failed.")
	}
	debugEnable, err := g.Cfg("customizeConfig").Get(gctx.New(), "customize.debugEnable")
	if err != nil {
		panic("get debug enable failed.")
	}
	initErrorLogger()
	initInfoLogger()
	// 异步日志
	initAsyncDebugLogger()

	EnableDebug(debugEnable.Bool())
	// 自定义handler类型的日志
	initHandlerLogger()

	initCriticalLogger()

	// 实现writer接口来自定义log操作
	warningLogger = &MyLoggerWriter{initWarningLog()}

	prodLogger = g.Log("prodLogger")
	devLogger = g.Log("devLogger")

	g.Config()
}

func SetLogger(loggerType string) {
	switch loggerType {
	case PROD:
		logger = prodLogger
	case DEV:
		logger = devLogger
	case INFO:
		logger = infoLogger
	case ERROR:
		logger = errorLogger
	case NOTICE:
		logger = customizeLogger
	case DEBUG:
		logger = asyncDebugLogger
	case WARN:
		myLogger = warningLogger
	default:
		logger = devLogger
	}
}

func PrintLog(ctx g.Ctx, level, message string) {
	SetLogger(level)
	if myLogger != nil {
		myLogger.Print(ctx, message)
		return
	} else if logger != nil {
		logger.Print(ctx, message)
		return
	}
	panic("please init log ")
}

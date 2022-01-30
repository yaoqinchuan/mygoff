package utils

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
)

var prodLogger *glog.Logger
var devLogger *glog.Logger
var infoLogger *glog.Logger
var errorLogger *glog.Logger
var handlerLogger *glog.Logger
var asyncDebugLogger *glog.Logger
var writerLogger *MyLoggerWriter

// 使用配置的方式來初始化錯誤日誌
func initErrorLogger() {
	errorLogger = glog.New()
	loggerPath := "logs/error"
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

// 可以打印栈信息
func initCollectLogger() {
	errorLogger = glog.New()
	loggerPath := "logs/critical"
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

// 链式初始化,
func initInfoLogger() {
	infoLogger = glog.New()
	// 链式初始化是没法改变infoLogger的配置的
	infoLogger.LevelStr("INFO").Path("logs/info").Stdout(true).File("{Y-m-d}[info].log").Line(true).Skip(1).Print(nil, "init info logger success.")
	fmt.Println(*infoLogger)
	infoLogger.SetLevel(glog.LEVEL_INFO)
	_ = infoLogger.SetPath("logs/info")
	infoLogger.SetStdoutPrint(true)
	infoLogger.SetFile("{Y-m-d}[info].log")
	infoLogger.SetStackSkip(1)
	infoLogger.SetWriterColorEnable(true)
}
func initCustomerLog() *glog.Logger {
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
	w.SetFile("{Y-m-d}[notice].log")
	err = w.SetPath("logs/notice")
	if err != nil {
		fmt.Println("set notice log path failed, error:" + err.Error())
		return nil
	}
	return w
}

// 自定义handler的日志，可以实现群通知和ES收集
func initHandlerLogger() {
	handlerLogger = initCustomerLog()
	handlerLogger.SetHandlers(loggingJsonHandler)
}

func initAsyncDebugLogger() {
	asyncDebugLogger = glog.New()
	asyncDebugLogger.LevelStr("DEBUG").Path("logs/debug").Stdout(true).File("{Y-m-d}[debug].log").Line(true).Async(true).SetWriterColorEnable(true)
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
	w.SetFile("{Y-m-d}[warn].log")
	err = w.SetPath("logs/warn")
	if err != nil {
		fmt.Println("set warning log path failed, error:" + err.Error())
		return nil
	}
	return w
}

// 选择是否开启debug模式
func enableDebug(flag bool) {
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
	writerLogger = &MyLoggerWriter{initWarningLog()}

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
func GetWriterLogger() *MyLoggerWriter {
	return writerLogger
}
// 自定义命令行,刚好为了测试日志才这样写的，实际上都应该初始化
var (
	logger   *glog.Logger
	myLogger *MyLoggerWriter

	mainCommand = &gcmd.Command{
		Name:        "main",
		Brief:       "start http server",
		Description: "this is the command entry for starting your process",
	}

	prodCommand = &gcmd.Command{
		Name:        "prod",
		Brief:       "open prod log mode",
		Description: "this is the command entry for opening prod log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetProdLogger()
			return
		},
	}

	devCommand = &gcmd.Command{
		Name:        "dev",
		Brief:       "open dev log mode",
		Description: "this is the command entry for opening dev log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetDevLogger()
			return
		},
	}

	errorCommand = &gcmd.Command{
		Name:        "error",
		Brief:       "open critical log mode",
		Description: "this is the command entry for opening critical log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetErrorLogger()
			return
		},
	}

	infoCommand = &gcmd.Command{
		Name:        "info",
		Brief:       "open info log mode",
		Description: "this is the command entry for opening information log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetInfoLogger()
			return
		},
	}

	customerCommand = &gcmd.Command{
		Name:        "customer",
		Brief:       "open customer log mode",
		Description: "this is the command entry for opening customer log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetHandlerLogger()
			return
		},
	}

	warningCommand = &gcmd.Command{
		Name:        "warn",
		Brief:       "open warn log mode",
		Description: "this is the command entry for opening warn log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			myLogger = GetWriterLogger()
			return
		},
	}

	debugCommand = &gcmd.Command{
		Name:        "debug",
		Brief:       "open debug log mode",
		Description: "this is the command entry for opening debug log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			logger = GetAsyncDebugLogger()
			enableDebug(true)
			return
		},
	}
)
func init() {
	err := mainCommand.AddCommand(prodCommand, devCommand, customerCommand, errorCommand, warningCommand, debugCommand, infoCommand)
	if err != nil {
		panic(err)
	}
}

func GetLoggerHandler() *gcmd.Command {
	return mainCommand
}

func PrintLog(ctx g.Ctx, message string) {
	if myLogger != nil {
		myLogger.Print(ctx, message)
		return
	}  else if logger != nil {
		logger.Print(ctx, message)
		return
	}
	panic("please init log ")
}


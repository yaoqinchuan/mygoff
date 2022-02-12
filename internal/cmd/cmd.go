package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"mygogf/internal/utils"
)


// 自定义命令行,刚好为了测试日志才这样写的，实际上都应该初始化
var (

	LoggerMode = utils.INFO

	MainCommand = &gcmd.Command{
		Name:        "main",
		Brief:       "start http server",
		Description: "this is the command entry for starting your process",
	}

	prodCommand = &gcmd.Command{
		Name:        "prod",
		Brief:       "open prod log mode",
		Description: "this is the command entry for opening prod log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.PROD
			return
		},
	}

	devCommand = &gcmd.Command{
		Name:        "dev",
		Brief:       "open dev log mode",
		Description: "this is the command entry for opening dev log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.DEV
			return
		},
	}

	errorCommand = &gcmd.Command{
		Name:        "error",
		Brief:       "open error log mode",
		Description: "this is the command entry for opening error log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.ERROR
			return
		},
	}

	infoCommand = &gcmd.Command{
		Name:        "info",
		Brief:       "open info log mode",
		Description: "this is the command entry for opening information log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.INFO
			return
		},
	}

	customizeCommand = &gcmd.Command{
		Name:        "customize",
		Brief:       "open customer log mode",
		Description: "this is the command entry for opening customer log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.NOTICE
			return
		},
	}

	warningCommand = &gcmd.Command{
		Name:        "warn",
		Brief:       "open warn log mode",
		Description: "this is the command entry for opening warn log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.WARN
			return
		},
	}

	debugCommand = &gcmd.Command{
		Name:        "debug",
		Brief:       "open debug log mode",
		Description: "this is the command entry for opening debug log mode",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			LoggerMode = utils.DEBUG
			return
		},
	}
)

func init() {
	err := MainCommand.AddCommand(prodCommand, devCommand, customizeCommand, errorCommand, warningCommand, debugCommand, infoCommand)
	if err != nil {
		panic(err)
	}
}

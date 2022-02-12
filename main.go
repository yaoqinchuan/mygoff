package main

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"mygogf/apiv1"
	"mygogf/errors"
	"mygogf/internal/cmd"
	"mygogf/internal/controller"
	_ "mygogf/internal/packed"
	"mygogf/internal/service/middleware"
	"mygogf/internal/utils"
)

// 从环境变量获取配置
func printEnv() error {
	v, err := g.Cfg().GetWithEnv(gctx.New(), "GOPATH")
	if err != nil || v == nil {
		return gerror.NewCode(errors.NotConfigured, "env GOPATH is not configured")
	}
	fmt.Printf("env GOPATH:%s\n", v)
	return nil
}

func initServer(ctx context.Context) {
	s := g.Server()
	s.BindMiddlewareDefault(middleware.ResponseHandler)
	apiv1.RegisterHelloController(ctx, s)
	apiv1.RegisterTotalController(ctx, s)
	apiv1.RegisterReflectController(ctx, s)
	controller.RegisterUserController(ctx, s)
	s.Run()
}

func main() {
	appCtx := gctx.New()
	err := printEnv()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cmd.MainCommand.Run(appCtx)
	utils.PrintLog(appCtx, utils.INFO, "app using "+cmd.LoggerMode+" level mode.")
	utils.PrintLog(appCtx, utils.INFO, "start application mygogf.")
	initServer(appCtx)
}

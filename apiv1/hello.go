package apiv1

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"mygogf/internal/utils"
)

type HelloReq struct {
	g.Meta `path:"/hello" tags:"Hello" method:"get" summary:"You first hello api"`
	Name   string `v:"required" dc:"your name"`
}
type HelloRes struct {
	Reply string `dc:"replay content"`
}
type Hello struct {
}

func (hello *Hello) Say(ctx context.Context, req *HelloReq) (res *HelloRes, err error) {
	utils.PrintLog(ctx, utils.INFO, fmt.Sprintf(`receive say: %v`, req))
	res = &HelloRes{Reply: fmt.Sprintf(`Hi %s`, req.Name)}
	return
}

func RegisterHelloController(ctx context.Context, server *ghttp.Server) {
	var hello = Hello{}
	server.BindObject("/hello", hello, "Say")
}

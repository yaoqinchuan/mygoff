package apiv1

/*
函数注册方式是最简单且最灵活的的路由注册方式，注册的服务可以是一个实例化对象的方法地址，也可以是一个包方法地址。
服务需要的数据可以通过模块内部变量形式或者对象内部变量形式进行管理，开发者可根据实际情况进行灵活控制。相关方法
func (s *Server) BindHandler(pattern string, handler interface{})
那么可以使用三种方式来实现这个注册：
 1、匿名函数注册
 2、有名函数注册
 3、对象方法注册
 本demo就使用的对象注册
*/

import (
	"github.com/gogf/gf/v2/container/gtype"
	"github.com/gogf/gf/v2/net/ghttp"
)
import (
	"context"
)

var (
	c = &TotalController{
		total: gtype.NewInt(),
	}
)

type TotalController struct {
	total *gtype.Int
}

func (totalController *TotalController) getTotal(r *ghttp.Request) {
	r.Response.Write("total:", totalController.total.Add(1))
}
//无法区分方法
func RegisterTotalController(ctx context.Context, server *ghttp.Server) {
	server.BindHandler("/total", c.getTotal)
}

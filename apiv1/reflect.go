package apiv1

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/net/ghttp"
	"mygogf/internal/utils"
	"time"
)

/*
对象注册注册一个实例化的对象，以后每一个请求都交给该对象（同一对象）处理，该对象常驻内存不释放。
func (s *Server) BindObject(pattern string, object interface{}, methods ...string) error
func (s *Server) BindObjectMethod(pattern string, object interface{}, method string) error
func (s *Server) BindObjectRest(pattern string, object interface{}) error
*/

type Reflect struct {
	cache string
}

func (reflect *Reflect) Index(request *ghttp.Request) {
	request.Response.Write("index")
}

func (reflect *Reflect) Show(request *ghttp.Request) {
	result := request.URL.Query().Get("data")
	if result == "" {
		request.Response.Write("empty context")
	} else {
		request.Response.Write(result)
	}
}

// restful 风格的uri与绑定对象不一样，这个需要注意了！
func (reflect *Reflect) Get(request *ghttp.Request) {
	request.Response.Write(reflect.cache)
}

func (reflect *Reflect) Post(request *ghttp.Request) {
	result := request.URL.Query().Get("cache")
	if result == "" {
		request.Response.Write("empty cache,please check.")
	} else {
		reflect.cache = result
		request.Response.Write(result)
	}
}

func (reflect *Reflect) Delete(request *ghttp.Request) {
	reflect.cache = "null"
	request.Response.Write(reflect.cache)
}

// 对象中的Init和Shut是两个在HTTP请求流程中被Server自动调用的特殊方法（类似构造函数和析构函数的作用）,添加处理开始时间。
func (reflect *Reflect) Init(request *ghttp.Request) {
	request.Response.Writeln("receive time:" + time.Now().String())
}

func (reflect *Reflect) Shut(request *ghttp.Request) {
	request.Response.Writeln("\nprocess end time:" + time.Now().String())
}
func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func MiddlewareLog(r *ghttp.Request) {
	r.Middleware.Next()
	utils.PrintLog(r.Context(), utils.INFO, fmt.Sprintf("url: %s, status: %s", r.URL.Path, r.Response.Status))
}
// 注册路由可以使用map group.ALLMap，但是不支持restful分割
func RegisterReflectController(ctx context.Context, server *ghttp.Server) {
	var reflectController = Reflect{
		"null",
	}
	/*
		1、 server.BindObject("/reflect", Reflect) 使用的是默认规格 /reflect/{method}
		  :8000   | ALL    | /reflect/delete | mygogf/apiv1.(*Reflect).Delete                                  |
		----------|--------|-----------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | /reflect/get    | mygogf/apiv1.(*Reflect).Get                                     |
		----------|--------|-----------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | /reflect/index  | mygogf/apiv1.(*Reflect).Index                                   |
		----------|--------|-----------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | /reflect/post   | mygogf/apiv1.(*Reflect).Post                                    |
		2、 server.BindObject("/ {.struct}/{.method}", reflectController) 当使用BindObject方法进行对象注册时，在路由规则中可以使用两个内置的变量：{.struct}和{.method}，前者表示当前对象名称，后者表示当前注册的方法名。
		  ADDRESS | METHOD |      ROUTE       |                             HANDLER                             |    MIDDLEWARE
		----------|--------|------------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | / reflect/delete | mygogf/apiv1.(*Reflect).Delete                                  |
		----------|--------|------------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | / reflect/get    | mygogf/apiv1.(*Reflect).Get                                     |
		----------|--------|------------------|-----------------------------------------------------------------|--------------------
		  :8000   | ALL    | / reflect/index  | mygogf/apiv1.(*Reflect).Index                                   |
		----------|--------|------------------|-----------------------------------------------------------------|-------------------
		3、 也可以使用restful风格的对象绑定
			 ADDRESS | METHOD |   ROUTE    |                             HANDLER                             |    MIDDLEWARE
			----------|--------|------------|-----------------------------------------------------------------|--------------------
			  :8000   | DELETE | / reflect  | mygogf/apiv1.(*Reflect).Delete                                  |
			----------|--------|------------|-----------------------------------------------------------------|--------------------
			  :8000   | GET    | / reflect  | mygogf/apiv1.(*Reflect).Get                                     |
			----------|--------|------------|-----------------------------------------------------------------|--------------------
			  :8000   | POST   | / reflect  | mygogf/apiv1.(*Reflect).Post                                    |
			----------|--------|------------|-----------------------------------------------------------------|--------------------
	*/
	server.BindObjectRest("/reflect", reflectController)
}

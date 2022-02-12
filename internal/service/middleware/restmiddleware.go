package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"mygogf/internal/utils"
)

func ResponseHandler(r *ghttp.Request) {
	defer func() {
		err := r.GetError()
		if err != nil {
			utils.PrintLog(r.GetCtx(), utils.ERROR, err.Error())
			r.Response.ClearBuffer()
			r.Response.Writef("%+v", err)
		}
	}()
	r.Middleware.Next()
}

package controller

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"mygogf/errors"
	"mygogf/internal/model/entity"
	"mygogf/internal/service/manager"
	"mygogf/internal/utils"
	"net/http"
	"strconv"
)

type UserController struct {
	g.Meta `path:"/user" method:"get,post,delete" dc:"user operate" tags:"User"`
}

func init() {
	// 全局注册
	// entity.RegisterUserCosmicValidator()
}

func (user *UserController) Get(request *ghttp.Request) {
	var ctx = request.Request.Context()
	ctx, span := gtrace.NewSpan(ctx, "GetUser")
	defer span.End()
	id := request.URL.Query().Get("id")
	if id == "" {
		request.Response.Write(gerror.NewCode(errors.NotFound, "query user by id failed, error: id is empty."))
		return
	}
	userId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		request.Response.WriteStatus(http.StatusInternalServerError, errors.InternalError.SetErrorContent(err.Error()))
		return
	}

	result, err := manager.FindUserByIdUsingGDB(ctx, userId)
	if err != nil {
		request.Response.Write(gerror.NewCode(errors.SQLError, "query user by id failed, error: "+err.Error()))
		return
	}
	request.Response.Write(result)
}

func (user *UserController) Post(request *ghttp.Request) {
	var input entity.User
	ctx := request.Request.Context()
	err := request.Parse(&input)
	if err != nil {
		request.Response.WriteStatus(http.StatusOK, errors.WrongParameterError.SetErrorContent(err.Error()))
		return
	}

	err = entity.UserCosmicValidator().Data(&input).Run(ctx)
	if err != nil {
		request.Response.WriteStatus(http.StatusOK, errors.WrongParameterError.SetErrorContent(err.Error()))
		return
	}
	if input.Id == 0 {
		err := manager.InsertUserUsingGDB(ctx, &input)
		if err != nil {
			request.Response.Write(gerror.NewCode(errors.DataTypeConvertError, "insert user failed, error: "+err.Error()))
			return
		}
	} else {
		result, err := manager.FindUserByIdUsingGDB(ctx, input.Id)
		if err != nil {
			request.Response.Write(gerror.NewCode(errors.NotFound, "query user when update or insert user failed, error: "+err.Error()))
			return
		}
		if result == nil{
			err := manager.InsertUserUsingGDB(ctx, &input)
			if err != nil {
				request.Response.Write(gerror.NewCode(errors.DataTypeConvertError, "insert user failed, error: "+err.Error()))
				return
			}
		} else {
			err := manager.UpdateUserUsingGDB(ctx, &input)
			if err != nil {
				request.Response.Write(gerror.NewCode(errors.DataTypeConvertError, "update user failed, error: "+err.Error()))
				return
			}
		}
	}

	request.Response.WriteStatus(http.StatusOK, errors.OK)
}

func (user *UserController) Delete(request *ghttp.Request) {
	id := request.URL.Query().Get("id")
	if id == "" {
		request.Response.Write(gerror.NewCode(errors.NotFound, "query user by id failed, error: id is empty."))
		return
	}
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		request.Response.Write(gerror.NewCode(errors.DataTypeConvertError, "query user by id failed, error: "+err.Error()))
		return
	}
	ctx := request.Request.Context()
	if userId == 0 {
		request.Response.Write(gerror.NewCode(errors.WrongParameterError, "delete user by id failed, error: id is empty"))
	} else {
		err = manager.DeleteUserByIdUsingGDB(ctx, userId)
	}
	if err != nil {
		utils.PrintLog(ctx, utils.ERROR, "delete user by id failed, error: "+err.Error())
		request.Response.Write(gerror.NewCode(errors.WrongParameterError, "delete user by id failed, error: "+err.Error()))
	}
}

func RegisterUserController(ctx context.Context, server *ghttp.Server) {
	var userController = UserController{}
	server.BindObjectRest("/user", userController)
}

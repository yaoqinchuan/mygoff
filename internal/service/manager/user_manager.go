package manager

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/gtime"
	"mygogf/internal/model/entity"
	"mygogf/internal/service/internal/dao"
	"mygogf/internal/service/internal/do"
)

func FindUserByIdUsingGDB(ctx context.Context, id uint64) (*entity.User, error) {
	resultDo, err := dao.FindUserByIdUsingGDB(ctx, id)
	if err != nil {
		return nil, err
	}
	if resultDo == nil {
		return nil, nil
	}
	result := &entity.User{
		Id:       id,
		Password: fmt.Sprintf(`%s`, resultDo.Password),
		Passport: fmt.Sprintf(`%s`, resultDo.Passport),
		Nickname: fmt.Sprintf(`%s`, resultDo.Nickname),
		CreateAt: resultDo.CreateAt,
		UpdateAt: resultDo.UpdateAt,
	}
	return result, nil
}

func InsertUserUsingGDB(ctx context.Context, input *entity.User) error {
	userDo := do.User{
		Id:       input.Id,
		Password: input.Password,
		Passport: input.Passport,
		Nickname: input.Nickname,
		CreateAt: gtime.Now(),
		UpdateAt: gtime.Now(),
	}
	return dao.InsertUserUsingGDB(ctx, &userDo)
}

func UpdateUserUsingGDB(ctx context.Context, input *entity.User) error {
	userDo := do.User{
		Id:       input.Id,
		Password: input.Password,
		Passport: input.Passport,
		Nickname: input.Nickname,
		CreateAt: input.CreateAt,
		UpdateAt: gtime.Now(),
	}
	return dao.UpdateUserUsingGDB(ctx, &userDo)
}

func DeleteUserByIdUsingGDB(ctx context.Context, id int64) error {
	return dao.DeleteUserByIdUsingGDB(ctx, id)
}

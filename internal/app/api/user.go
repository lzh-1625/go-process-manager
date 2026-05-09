package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

type userApi struct{}

var UserApi = new(userApi)

func (u *userApi) LoginHandler(ctx *echo.Context) error {
	var req model.LoginHandlerReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	user, ok := logic.UserLogic.CheckLoginInfo(req.Account, req.Password)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, model.Response[struct{}]{
			Message: "incorrect username or password",
			Code:    -1,
		})
	}
	token, err := utils.GenerateToken(req.Account)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[map[string]any]{
		Message: "login success",
		Code:    0,
		Data: map[string]any{
			"token":    token,
			"username": req.Account,
			"role":     user.Role,
		},
	})
}

func (u *userApi) CreateUser(ctx *echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = logic.UserLogic.CreateUser(req)
	return
}

func (u *userApi) EditUser(ctx *echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = logic.UserLogic.EditUser(req, getUserName(ctx), getRole(ctx))
	return
}

func (u *userApi) DeleteUser(ctx *echo.Context) (err error) {
	var req struct {
		Account string `query:"account"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = logic.UserLogic.DeleteUser(req.Account)
	return
}

func (u *userApi) GetUserList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]*model.User]{
		Data:    logic.UserLogic.GetUserList(),
		Message: "success",
		Code:    0,
	})
}

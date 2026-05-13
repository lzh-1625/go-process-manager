package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

type UserApi struct {
	userLogic *logic.UserLogic
}

func NewUserApi(userLogic *logic.UserLogic) *UserApi {
	return &UserApi{
		userLogic: userLogic,
	}
}

func (u *UserApi) LoginHandler(ctx *echo.Context) error {
	var req model.LoginHandlerReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	user, ok := u.userLogic.CheckLoginInfo(req.Account, req.Password)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, model.Response[struct{}]{
			Message: "incorrect username or password",
			Code:    -1,
		})
	}
	token, err := utils.GenerateToken(req.Account, config.CF.SecretKey, time.Now().Add(time.Duration(config.CF.TokenExpirationTime)*time.Hour))
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

func (u *UserApi) CreateUser(ctx *echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = u.userLogic.CreateUser(req)
	return
}

func (u *UserApi) EditUser(ctx *echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = u.userLogic.EditUser(req, getUserName(ctx), getRole(ctx))
	return
}

func (u *UserApi) DeleteUser(ctx *echo.Context) (err error) {
	var req struct {
		Account string `query:"account"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	err = u.userLogic.DeleteUser(req.Account)
	return
}

func (u *UserApi) GetUserList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]*model.User]{
		Data:    u.userLogic.GetUserList(),
		Message: "success",
		Code:    0,
	})
}

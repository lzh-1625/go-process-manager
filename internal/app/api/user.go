package api

import (
	"errors"

	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"
)

type userApi struct{}

var UserApi = new(userApi)

const DEFAULT_ROOT_PASSWORD = "root"

func (u *userApi) LoginHandler(ctx echo.Context) error {
	var req model.LoginHandlerReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if !u.checkLoginInfo(req.Account, req.Password) {
		return ctx.JSON(401, model.Response[struct{}]{
			Message: "incorrect username or password",
			Code:    -1,
		})
	}
	token, err := utils.GenerateToken(req.Account)
	if err != nil {
		return err
	}
	return ctx.JSON(200, model.Response[map[string]any]{
		Message: "login success",
		Code:    0,
		Data: map[string]any{
			"token":    token,
			"username": req.Account,
			"role":     repository.UserRepository.GetUserByName(req.Account).Role,
		},
	})
}

func (u *userApi) CreateUser(ctx echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if req.Role == eum.RoleRoot {
		return errors.New("creation of root accounts is forbidden")
	}
	if req.Account == eum.Console {
		return errors.New("operation failed")
	}
	if len(req.Password) < config.CF.UserPassWordMinLength {
		return errors.New("password is too short")
	}
	err = repository.UserRepository.CreateUser(req)
	return
}

func (u *userApi) EditUser(ctx echo.Context) (err error) {
	var req model.User
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	reqUser := getUserName(ctx)
	if req.Account == "root" {
		return errors.New("can not edit root user")
	}
	if getRole(ctx) != eum.RoleRoot && req.Account != "" {
		return errors.New("invalid parameters")
	}
	if req.Account == "" {
		req.Account = reqUser
	}
	if len(req.Password) != 0 && len(req.Password) < config.CF.UserPassWordMinLength {
		return errors.New("password is too short")
	}
	err = repository.UserRepository.EditUser(req)
	return
}

func (u *userApi) DeleteUser(ctx echo.Context) (err error) {
	var req struct {
		Account string `form:"account"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	if req.Account == "root" {
		return errors.New("deletion of root accounts is forbidden")
	}
	err = repository.UserRepository.DeleteUser(req.Account)
	return
}

func (u *userApi) GetUserList(ctx echo.Context) error {
	return ctx.JSON(200, model.Response[[]*model.User]{
		Data:    repository.UserRepository.GetUserList(),
		Message: "success",
		Code:    0,
	})
}

func (u *userApi) checkLoginInfo(account, password string) bool {
	user := repository.UserRepository.GetUserByName(account)
	if user == nil && account == "root" { // 如果root用户不存在，则创建一个root用户
		repository.UserRepository.CreateUser(model.User{
			Account:  "root",
			Password: DEFAULT_ROOT_PASSWORD,
			Role:     eum.RoleRoot,
		})
		return password == DEFAULT_ROOT_PASSWORD
	}
	return user != nil && user.Password == utils.Md5(password)
}

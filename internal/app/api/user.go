package api

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"

	"github.com/gin-gonic/gin"
)

type userApi struct{}

var UserApi = new(userApi)

const DEFAULT_ROOT_PASSWORD = "root"

func (u *userApi) LoginHandler(ctx *gin.Context, req model.LoginHandlerReq) any {
	if !u.checkLoginInfo(req.Account, req.Password) {
		return errors.New("incorrect username or password")
	}
	token, err := utils.GenerateToken(req.Account)
	if err != nil {
		return err
	}
	return gin.H{
		"token":    token,
		"username": req.Account,
		"role":     repository.UserRepository.GetUserByName(req.Account).Role,
	}
}

func (u *userApi) CreateUser(ctx *gin.Context, req model.User) (err error) {
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

func (u *userApi) EditUser(ctx *gin.Context, req model.User) (err error) {
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

func (u *userApi) DeleteUser(ctx *gin.Context, req struct {
	Account string `form:"account"`
}) (err error) {
	if req.Account == "root" {
		return errors.New("deletion of root accounts is forbidden")
	}
	err = repository.UserRepository.DeleteUser(req.Account)
	return
}

func (u *userApi) GetUserList(ctx *gin.Context, _ any) any {
	return repository.UserRepository.GetUserList()
}

func (u *userApi) checkLoginInfo(account, password string) bool {
	user := repository.UserRepository.GetUserByName(account)
	if account == "root" && user.Account == "" { // 如果root用户不存在，则创建一个root用户
		repository.UserRepository.CreateUser(model.User{
			Account:  "root",
			Password: DEFAULT_ROOT_PASSWORD,
			Role:     eum.RoleRoot,
		})
		return password == DEFAULT_ROOT_PASSWORD
	}
	return user.Password == utils.Md5(password)
}

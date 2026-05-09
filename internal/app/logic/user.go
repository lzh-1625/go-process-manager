package logic

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"
)

type userLogic struct{}

var UserLogic = new(userLogic)

const DefaultRootAccount = "root"
const DefaultRootPassword = "root"

func (u *userLogic) CheckLoginInfo(account, password string) (*model.User, bool) {
	user := repository.UserRepository.GetUserByName(account)
	if user == nil && account == DefaultRootAccount {
		user = &model.User{
			Account:  DefaultRootAccount,
			Password: DefaultRootPassword,
			Role:     eum.RoleRoot,
		}
		if err := repository.UserRepository.CreateUser(*user); err != nil {
			return nil, false
		}
		return user, password == DefaultRootPassword
	}
	return user, user != nil && user.Password == utils.Md5(password)
}

func (u *userLogic) CreateUser(user model.User) error {
	if user.Role == eum.RoleRoot {
		return errors.New("creation of root accounts is forbidden")
	}
	if user.Account == DefaultRootAccount {
		return errors.New("operation failed")
	}
	if len(user.Password) < config.CF.UserPassWordMinLength {
		return errors.New("password is too short")
	}
	return repository.UserRepository.CreateUser(user)
}

func (u *userLogic) EditUser(user model.User, currentAccount string, currentRole eum.Role) error {
	if user.Account == DefaultRootAccount {
		return errors.New("can not edit root user")
	}
	if currentRole != eum.RoleRoot && user.Account != "" {
		return errors.New("invalid parameters")
	}
	if user.Account == "" {
		user.Account = currentAccount
	}
	if len(user.Password) != 0 && len(user.Password) < config.CF.UserPassWordMinLength {
		return errors.New("password is too short")
	}
	return repository.UserRepository.EditUser(user)
}

func (u *userLogic) DeleteUser(account string) error {
	if account == DefaultRootAccount {
		return errors.New("deletion of root accounts is forbidden")
	}
	return repository.UserRepository.DeleteUser(account)
}

func (u *userLogic) GetUserList() []*model.User {
	return repository.UserRepository.GetUserList()
}

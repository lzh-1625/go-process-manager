package logic

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/utils"
)

type UserLogic struct {
	userRepository *repository.UserRepository
}

func NewUserLogic(userRepository *repository.UserRepository) *UserLogic {
	return &UserLogic{
		userRepository: userRepository,
	}
}

const DefaultRootAccount = "root"
const DefaultRootPassword = "root"

func (u *UserLogic) CheckLoginInfo(account, password string) (*model.User, bool) {
	user := u.userRepository.GetUserByName(account)
	if user == nil && account == DefaultRootAccount {
		user = &model.User{
			Account:  DefaultRootAccount,
			Password: DefaultRootPassword,
			Role:     eum.RoleRoot,
		}
		if err := u.userRepository.CreateUser(*user); err != nil {
			return nil, false
		}
		return user, password == DefaultRootPassword
	}
	return user, user != nil && user.Password == utils.Md5(password)
}

func (u *UserLogic) CreateUser(user model.User) error {
	if user.Role == eum.RoleRoot {
		return errors.New("creation of root accounts is forbidden")
	}
	if user.Account == DefaultRootAccount {
		return errors.New("operation failed")
	}
	if len(user.Password) < config.CF.UserPassWordMinLength {
		return errors.New("password is too short")
	}
	return u.userRepository.CreateUser(user)
}

func (u *UserLogic) EditUser(user model.User, currentAccount string, currentRole eum.Role) error {
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
	return u.userRepository.EditUser(user)
}

func (u *UserLogic) DeleteUser(account string) error {
	if account == DefaultRootAccount {
		return errors.New("deletion of root accounts is forbidden")
	}
	return u.userRepository.DeleteUser(account)
}

func (u *UserLogic) GetUserList() []*model.User {
	return u.userRepository.GetUserList()
}

func (u *UserLogic) GetUserByName(name string) *model.User {
	return u.userRepository.GetUserByName(name)
}

package repository

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/utils"
)

type userRepository struct{}

var UserRepository = new(userRepository)

func (u *userRepository) GetUserByName(name string) *model.User {
	user, _ := query.User.Where(query.User.Account.Eq(name)).First()
	return user
}

func (u *userRepository) CreateUser(user model.User) error {
	user.Password = utils.Md5(user.Password)
	user.CreateTime = time.Now()
	return query.User.Create(&user)
}

func (u *userRepository) EditUser(data model.User) error {
	data.Password = utils.Md5(data.Password)
	_, err := query.User.Where(query.User.Account.Eq(data.Account)).Updates(&data)
	return err
}

func (u *userRepository) DeleteUser(name string) error {
	_, err := query.User.Where(query.User.Account.Eq(name)).Delete()
	return err
}

func (u *userRepository) GetUserList() (result []*model.User) {
	result, _ = query.User.Find()
	return
}

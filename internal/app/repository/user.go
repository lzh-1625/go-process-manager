package repository

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
	"github.com/lzh-1625/go_process_manager/utils"
)

func NewUserRepository() *UserRepository {
	return &UserRepository{
		query: query.Q,
	}
}

type UserRepository struct {
	query *query.Query
}

func (u *UserRepository) GetUserByName(name string) *model.User {
	user, _ := u.query.User.Where(u.query.User.Account.Eq(name)).First()
	return user
}

func (u *UserRepository) CreateUser(user model.User) error {
	user.Password = utils.Md5(user.Password)
	user.CreateTime = time.Now()
	return u.query.User.Create(&user)
}

func (u *UserRepository) EditUser(data model.User) error {
	data.Password = utils.Md5(data.Password)
	_, err := u.query.User.Where(u.query.User.Account.Eq(data.Account)).Updates(&data)
	return err
}

func (u *UserRepository) DeleteUser(name string) error {
	_, err := u.query.User.Where(u.query.User.Account.Eq(name)).Delete()
	return err
}

func (u *UserRepository) GetUserList() (result []*model.User) {
	result, _ = u.query.User.Find()
	return
}

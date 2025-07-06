package repository

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

type userRepository struct{}

var UserRepository = new(userRepository)

func (u *userRepository) GetUserByName(name string) model.User {
	var result model.User
	db.Model(&model.User{}).Where(&model.User{Account: name}).First(&result)
	return result
}

func (u *userRepository) CreateUser(user model.User) error {
	user.Password = utils.Md5(user.Password)
	user.CreateTime = time.Now()
	tx := db.Create(&user)
	return tx.Error
}

func (u *userRepository) UpdatePassword(name string, password string) error {
	return db.Model(&model.User{}).Where(&model.User{Account: name}).Updates(&model.User{Password: utils.Md5(password)}).Error
}

func (u *userRepository) DeleteUser(name string) error {
	return db.Delete(&model.User{Account: name}).Error
}

func (u *userRepository) GetUserList() (result []model.User) {
	db.Find(&result)
	return
}

package service

import (
	"chatapp/base/appconfig"
	"chatapp/dao"
	"chatapp/objects"
	"fmt"
)

type UserService interface {
	Save(objects.User) *objects.User
	FindAll(offset int,limit int) []*objects.User
	FindById(id int) (*objects.User, error)
	UpdateById(int, objects.User) (*objects.User, error)
	FindByUsernameAndPassword(username string, password string) (*objects.User, error)
	DeleteById(int) error
}

type userService struct {
	userDao dao.UserDao
	config *appconfig.Config
}

func (service *userService) Save(user objects.User) *objects.User{
	id, err := service.userDao.Create(user)
	if err != nil {
		fmt.Println("user save error: ", err)
		return nil
	}
	user.Id = id
	return &user
}

func (service *userService) FindAll(offset int, limit int) []*objects.User {
	users,_,err := service.userDao.QueryAll(offset, limit)
	if err != nil {
		fmt.Println("queryall error: ", err)
	}
	return users
}

func (service *userService) FindById(id int) (*objects.User, error) {
	user, err := service.userDao.FindById(id)
	if err != nil {
		fmt.Println("FindById error: ", err)
		return nil, err 
	}

	return user, nil 
}

func (service *userService) UpdateById(id int, user objects.User) (*objects.User, error) {
	newUser, err := service.userDao.UpdateById(id, user)
	if err != nil {
		fmt.Println("UpdateById error: ", err)
		return nil, err 
	}

	return newUser, nil 
}


func (service *userService) DeleteById(id int) error {
	err := service.userDao.DeleteById(id)
	if err != nil {
		fmt.Println("DeleteById error: ", err)
		return err
	}
	return nil
}


func (service *userService) FindByUsernameAndPassword(username string, password string) (*objects.User, error) {
	user, err := service.userDao.FindByUsernameAndPassword(username, password)
	if err != nil {
		fmt.Println("DeleteById error: ", err)
		return nil, err
	}
	return user, nil
}

func NewUserService(config *appconfig.Config) UserService {
	return &userService{userDao : dao.NewUserDao(config), config: config}
}

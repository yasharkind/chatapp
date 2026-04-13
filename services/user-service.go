package service

import (
	"chatapp/base/appconfig"
	"chatapp/dao"
	"chatapp/objects"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserService interface {
	Save(objects.User) *objects.User
	FindAll(offset int,limit int) []*objects.User
	FindById(id int) (*objects.User, error)
	UpdateById(int, objects.User) (*objects.User, error)
	FindByUsernameAndPassword(username string, password string) (*objects.User, error)
	DeleteById(int) error
	JWTByUsernameAndPassword(string, string) (string, error)
	Validate(cookie string) *objects.User
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

func (service *userService) JWTByUsernameAndPassword(username string, password string) (string, error) {
	user, err := service.FindByUsernameAndPassword(username, password)
	if err != nil {
		fmt.Printf("jwtbyemailandpassword error: %s", err)
		return "", err
	}
	if user == nil {
		fmt.Printf("user by username %s not found", username)
		return "", nil
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(service.config.Secret))
	if err != nil {
		fmt.Printf("failed to create token %s", err)
		return "", err
	}
	return tokenString, nil
}


func (service *userService) Validate(cookie string) *objects.User {
		if cookie == "" {
			return nil
		}
		token, err := jwt.Parse(cookie, func(token *jwt.Token) (any, error) {
			if _,ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
			}
	
			return service.config.Secret, nil;
		})
		if token == nil {
			return nil
		}
		claims, _ := token.Claims.(jwt.MapClaims)

		if claims["sub"] == nil || claims["exp"] == nil {
			fmt.Println("token invalid", token)
			return nil
		}
	
		user, err := service.FindById(int(claims["sub"].(float64)))
		t2 := time.Unix(int64(claims["exp"].(float64)), 0)
		if err != nil {
			log.Printf("validate error %s", err)
		}
		expired := t2.Unix() <= time.Now().Unix() 
		if user != nil && !expired {
			return user
		}
		println(expired)
		return nil
	}

func NewUserService(config *appconfig.Config) UserService {
	return &userService{userDao : dao.NewUserDao(config), config: config}
}

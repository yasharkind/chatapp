package controller

import (
	"chatapp/base/appconfig"
	dto_in "chatapp/dto/in"
	dto_out "chatapp/dto/out"
	"chatapp/middlewares"
	service "chatapp/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type UserController struct {
	userService service.UserService 
	config *appconfig.Config
}

func (c *UserController) handleLogin(w http.ResponseWriter, r *http.Request){
	var user_dto *dto_in.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user_dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("decode error: ", err)
		return
	}
	token, err := c.userService.JWTByUsernameAndPassword(user_dto.Username, user_dto.Password)
	if err != nil {
		fmt.Println("jwt error: ", err)
		w.WriteHeader(403)
	}
	w.Write([]byte(token))
}

func (c *UserController) handleUser(w http.ResponseWriter, r *http.Request) {
	var token = strings.Join(r.Header["Authorization"], "")
	user := c.userService.Validate(token)
	if user == nil {
		w.WriteHeader(404)
		return
	}
	res := dto_out.User{
		Id: user.Id,
		Username: user.Username,
		Avatar: user.Avatar,
	}
	marshaled, err := json.Marshal(res)
	if err != nil {
		println(err)
		w.WriteHeader(401)
		return
	}
	w.Write(marshaled)
}


func NewUserController(config *appconfig.Config, userService service.UserService) *UserController {
	c := UserController{
		config: config,
		userService: userService,
	}

	http.Handle("/avatar/",  middlewares.CorsMiddleware(http.StripPrefix("/avatar/", http.FileServer(http.Dir("avatars")))))

	http.Handle("/login", middlewares.CorsMiddleware(http.HandlerFunc(c.handleLogin)))

	http.Handle("/user", middlewares.CorsMiddleware(http.HandlerFunc(c.handleUser)))

	return &c

}


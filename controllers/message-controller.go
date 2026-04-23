package controller

import (
	"chatapp/base/appconfig"
	dto_in "chatapp/dto/in"
	"chatapp/dto/in/opcodes"
	dto_out "chatapp/dto/out"
	"chatapp/middlewares"
	"chatapp/objects"
	service "chatapp/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)


type MessageController struct {
	config *appconfig.Config
	messageService service.MessageService
	userService service.UserService
	mu sync.Mutex
	broadcast func(dto_out.WebSockRes)
}

func (c *MessageController) handleMessageHistory(w http.ResponseWriter, r *http.Request) {
	token := strings.Join(r.Header["Authorization"], "")
	if (token == "") {
		fmt.Println("token is empty")
		return
	}
	user := c.userService.Validate(token)
	if user == nil {
		fmt.Println("user not found")
		return
	}
	c.mu.Lock()
	messageList := c.messageService.FindByLimitOffset(c.config.Server.MessageLimit, 0)
	c.mu.Unlock()
	
	var messageDtos []dto_out.SendMessage
	for _, message := range messageList {
		messageDtos = append(messageDtos, dto_out.SendMessage{
			Id: message.Id,
			Sender: message.Sender.Username,
			SenderId: message.Sender.Id,
			Content: message.Content,
			Avatar: message.Sender.Avatar,
			Time: message.Time.Local().Format("15:04:05"),
		})
	}
	dto := dto_out.Auth{
		Token: token,
		Messages: messageDtos,
	}
	marshaled, err := json.Marshal(dto)
	if err != nil {
		fmt.Println("marshal error: ", err)
	}
	if _, err := w.Write(marshaled); err != nil {
		fmt.Println("write error: ", err)
	}
}

func (c *MessageController) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	var token = strings.Join(r.Header["Authorization"], "")
	user := c.userService.Validate(token)
	if user == nil {
		w.WriteHeader(403)
		return
	}
	var dto dto_in.SendMessage
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&dto)
	msg := objects.Message{
		Id: 0,
		Time: time.Now().Local(),
		Content: dto.Content,
		Sender: user,
	}

	message := c.messageService.Save(msg)

	wsR := dto_out.SendMessage{
		Action: opcodes.SendMessage,
		Id: message.Id,
		Sender: user.Username,
		SenderId: user.Id,
		Content: message.Content,
		Avatar: user.Avatar,
		Time: message.Time.Local().Format("15:04:05"),
	}

	c.broadcast(&wsR)
}

func (c *MessageController) handleEditMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	var token = strings.Join(r.Header["Authorization"], "")
	user := c.userService.Validate(token)
	if user == nil {
		w.WriteHeader(403)
		return
	}
	var dto dto_in.EditMessage
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&dto)

	message, err := c.messageService.FindById(dto.MessageId)
	if err != nil{
		println("DeleteMessage error 1: ", err)
		w.WriteHeader(401)
		return
	}
	if (message.Sender.Id != user.Id) {
		println("DeleteMessage error 2: no permission ", message.Sender.Id, user.Id)
		w.WriteHeader(403)
		return
	}
	newmsg := objects.Message{
		Id: message.Id,
		Content: dto.NewContent,
		Sender: message.Sender,
		Time: message.Time,
	}
	msg, err := c.messageService.UpdateById(dto.MessageId, newmsg)
	if err != nil {
		println("DeleteMessage error 3: ", err)
		w.WriteHeader(401)
		return
	}
	response := &dto_out.EditMessage{
		Action: opcodes.EditMessage,
		Id: dto.MessageId,
		NewContent: msg.Content,
	}

	log := fmt.Sprintf("%s-%s: Edited Message %d content: %s", time.Now().Local(), user.Username, dto.MessageId, dto.NewContent)
	println(log)
	c.broadcast(response)
	w.WriteHeader(200)
}

func (c *MessageController) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(404)
		return
	}
	var token = strings.Join(r.Header["Authorization"], "")
	user := c.userService.Validate(token)
	if user == nil {
		w.WriteHeader(404)
		return
	}
	var dto dto_in.DeleteMessage
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&dto)

	message, err := c.messageService.FindById(dto.MessageId)
	if err != nil{
		println("DeleteMessage error 1: ", err)
		w.WriteHeader(401)
		return
	}
	if (message.Sender.Id != user.Id) {
		println("DeleteMessage error 2: no permission ", message.Sender.Id, user.Id)
		w.WriteHeader(403)
		return
	}
	err = c.messageService.DeleteById(dto.MessageId)
	if err != nil {
		println("DeleteMessage error 3: ", err)
		w.WriteHeader(401)
		return
	}
	response := &dto_out.DeleteMessage{
		Action: opcodes.DeleteMessage,
		Id: dto.MessageId,
	}

	log := fmt.Sprintf("%s-%s: Deleted Message %d", time.Now().Local(), user.Username, dto.MessageId)
	println(log)
	c.broadcast(response)
	w.WriteHeader(200)
}


func (c *MessageController) handleLoadMoreMessages(w http.ResponseWriter, r *http.Request) {
	token := strings.Join(r.Header["Authorization"], "")

	user := c.userService.Validate(token)

	if user == nil {
		w.WriteHeader(403)
		return
	}

	offset, err := strconv.Atoi(r.PathValue("offset"))

	if err != nil {
		w.WriteHeader(401)
		return
	}
	messages := c.messageService.FindByLimitOffset(c.config.Server.MessageLimit, offset * c.config.Server.MessageLimit)

	var dto []dto_out.SendMessage
	for _, message := range messages {
		obj := dto_out.SendMessage{
			Id: message.Id,
			Sender: message.Sender.Username,
			SenderId: message.Sender.Id,
			Content: message.Content,
			Avatar: message.Sender.Avatar,
			Time: message.Time.Local().Format("15:04:05") ,
		}
		dto = append(dto, obj)
	}

	marshaled, err := json.Marshal(dto)
	if err != nil {
		println("/messages marshal error: ", err)
	}

	w.Write(marshaled)
}



func NewMessageController (config *appconfig.Config, messageService service.MessageService, userService service.UserService, broadcast func(dto_out.WebSockRes)) *MessageController{
	c := MessageController{
		config: config,
		messageService: messageService,
		userService: userService,
		broadcast: broadcast,
	}
	http.Handle("/auth", middlewares.CorsMiddleware(http.HandlerFunc(c.handleMessageHistory)))
	http.Handle("/deletemessage", middlewares.CorsMiddleware(http.HandlerFunc(c.handleDeleteMessage)))
	http.Handle("/editmessage",  middlewares.CorsMiddleware(http.HandlerFunc(c.handleEditMessage)))
	http.Handle("GET /message/{offset}",  middlewares.CorsMiddleware(http.HandlerFunc(c.handleLoadMoreMessages)))
	http.Handle("PUT /message",  middlewares.CorsMiddleware(http.HandlerFunc(c.handleSendMessage)))
	http.Handle("OPTIONS /message/{offset}",  middlewares.CorsMiddleware(http.HandlerFunc(c.handleLoadMoreMessages)))

	return &c
}

package objects

import (
	"time"
)

type Message struct {
	Id int `json:"id"`
	Sender *User `json:"sender"`
	Time time.Time `json:"time"`
	Content string `json:"content"`
}

type User struct {
	Id int `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}

package objects

import (
	"time"
)

type Message struct {
	Id      int       `json:"id"`
	Sender  *User     `json:"sender"`
	Time    time.Time `json:"time"`
	Content string    `json:"content"`
}

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
}

type Denpa struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Url		  string `json:"url"`
	Singers   string `json:"singers"`
}

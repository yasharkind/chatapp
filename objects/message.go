package objects

import "time"

type Message struct {
	Id int `json:"id"`
	Sender string `json:"sender"`
	Time time.Time `json:"time"`
	Content string `json:"content"`
}

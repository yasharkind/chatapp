package dto

type Message struct {
	Sender string `json:"sender"`
	Password string `json:"password"`
	Content string `json:"content"`
}


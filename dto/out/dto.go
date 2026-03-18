package dto_out

import "chatapp/dto/in/opcodes"

type WebSockRes interface {
	OpCode() int
}


type SendMessage struct {
	Action 	 int    `json:"op_code"`
	Id       int    `json:"id"`
	Sender   string `json:"sender"`
	SenderId int    `json:"sender_id"`
	Content  string `json:"content"`
	Avatar   string `json:"avatar"`
	Time     string `json:"time"`
}

func (i *SendMessage) OpCode() int { return opcodes.SendMessage }


type DeleteMessage struct {
	Action int `json:"op_code"` // opcode for the frontend to handle
	Id	   int `json:"id"`
}

func (i *DeleteMessage) OpCode() int { return opcodes.DeleteMessage }


type User struct {
	Id       int   	`json:"id"`
	Username string `json:"username"`
	Avatar	 string `json:"avatar"`
}

type Auth struct {
	// OpCode int `json:"op_code"`
	Token	 string	   `json:"token"`
	Messages []SendMessage `json:"messages"`
}

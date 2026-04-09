package dto_in

type User struct {
	Id int `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Token string `json:"token"`
	OpCode int `json:"op_code"`
	Content string `json:"content"`
}

type DeleteMessage struct {
	MessageId int `json:"message_id"`
}

type EditMessage struct {
	MessageId  int    `json:"message_id"`
	NewContent string `json:"new_content"`
}

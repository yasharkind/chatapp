package dto_out


type Message struct {
	Id       int    `json:"id"`
	Sender   string `json:"sender"`
	//SenderId string `json:"sender_id"`
	Content  string `json:"content"`
	Avatar   string `json:"avatar"`
	Time     string `json:"time"`
}


type User struct {
	Id       int   	`json:"id"`
	Username string `json:"username"`
	Avatar	 string `json:"avatar"`
}

type Auth struct {
	// OpCode int `json:"op_code"`
	Token	 string	   `json:"token"`
	Messages []Message `json:"messages"`
}

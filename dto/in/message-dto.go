package dto_in

type Message struct {
	Token string `json:"token"`
	OpCode int `json:"op_code"`
	Content string `json:"content"`
}


package dto_out


type Message struct {
	Id      int    `json:"id"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Avatar  string `json:"avatar"`
	Time    string `json:"time"`
}

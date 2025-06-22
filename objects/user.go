package objects

type User struct {
	Id int `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}

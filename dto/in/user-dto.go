package dto_in

type User struct {
	Id int `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` // replace with JWT token later
}

package model

type User struct {
	ID       string `json:"-" db:"id"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

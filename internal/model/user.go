package model

type User struct {
	ID       string `json:"id" db:"id"`
	Key      string `json:"key" db:"key"`
	Email    string `json:"email" db:"email"`
	Password string `json:"_" db:"password"`
}

const (
	UserKey = "__user__"
)

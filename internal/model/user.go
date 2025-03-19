package model

type User struct {
	ID        string `json:"id" db:"id"`
	Key       string `json:"key" db:"key"`
	Email     string `json:"email" db:"email"`
	Password  string `json:"-" db:"password"`
	TimeGoalM int    `json:"-" db:"time_goal_m"`
}

const (
	UserKey = "__user_id__"
)

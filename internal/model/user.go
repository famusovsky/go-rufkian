package model

type User struct {
	ID        string  `json:"id" db:"id"`
	Email     string  `json:"email" db:"email"`
	Password  string  `json:"-" db:"password"`
	Key       *string `json:"key" db:"key"`
	TimeGoalM *int    `json:"-" db:"time_goal_m"`
}

func (u User) HasKey() bool {
	if u.Key == nil {
		return false
	}
	return len(*u.Key) > 0
}

func (u User) HasTimeGoal() bool {
	return u.TimeGoalM != nil
}

const (
	UserKey = "__user_id__"
)

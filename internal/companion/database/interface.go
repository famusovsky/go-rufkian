package database

import "github.com/famusovsky/go-rufkian/internal/model"

type IClient interface {
	AddUser(user model.User) (model.User, error)
	GetUser(id string) (model.User, error)
	GetUserByCredentials(user model.User) (model.User, error)
}

package database

import "github.com/famusovsky/go-rufkian/internal/model"

type IClient interface {
	AddUser(user model.User) (model.User, error)
	GetUser(id string) (model.User, error)
	GetUserByCredentials(user model.User) (model.User, error)

	GetDialog(id, userID string) (model.Dialog, error)
	GetUserDialogs(userID string) (model.Dialogs, error)

	AddWordToUser(userID, word string) error
	GetUserWords(userID string) ([]string, error)
	CheckUserWord(userID string, word string) (bool, error)
	DeleteWordFromUser(userID, word string) error
}

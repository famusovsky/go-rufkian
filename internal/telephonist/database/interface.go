package database

import "github.com/famusovsky/go-rufkian/internal/model"

type IClient interface {
	StoreDialog(userID uint64, messages model.Messages) (id uint64, err error)
}

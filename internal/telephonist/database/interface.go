package database

import "github.com/famusovsky/go-rufkian/internal/model"

type IClient interface {
	StoreDialog(key string, messages model.Messages) (id uint64, err error)
}

package database

import "github.com/famusovsky/go-rufkian/internal/model"

type IClient interface {
	StoreDialog(dialog model.Dialog) (model.Dialog, error)
}

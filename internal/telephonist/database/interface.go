package database

import "github.com/famusovsky/go-rufkian/internal/model"

//go:generate mockgen -package database -mock_names IClient=ClientMock -source ./interface.go -typed -destination interface.mock.gen.go
type IClient interface {
	StoreDialog(dialog model.Dialog) (model.Dialog, error)
	UpdateDialog(dialog model.Dialog) error
}

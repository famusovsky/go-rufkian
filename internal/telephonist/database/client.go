package database

import (
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type client struct {
	db     sqlx.Ext
	logger *zap.Logger
}

func NewClient(db sqlx.Ext, logger *zap.Logger) IClient {
	return client{
		db:     db,
		logger: logger,
	}
}

func (c client) StoreDialog(key string, messages model.Messages) (uint64, error) {
	messages = messages.Dialog()
	if len(messages) == 0 {
		c.logger.Error("store empty dialog", zap.String("key", key))
		return 0, model.ErrEmptyDialog
	}

	if c.db == nil {
		c.logger.Info("attempt to store dialog into nil db", zap.String("key", key), zap.Any("dialog", messages))
		return 0, nil
	}

	var id uint64
	if err := c.db.QueryRowx(storeDialogQuery, key, messages).Scan(&id); err != nil {
		c.logger.Error("store dialog sql query process", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	c.logger.Info("store dialog", zap.String("key", key), zap.Uint64("id", id))
	return id, nil
}

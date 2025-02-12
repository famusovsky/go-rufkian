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

func (c client) StoreDialog(userID uint64, messages model.Messages) (uint64, error) {
	messages = messages.Dialog()
	if len(messages) == 0 {
		c.logger.Error("store empty dialog", zap.Uint64("user_id", userID))
		return 0, model.ErrEmptyDialog
	}

	if c.db == nil {
		c.logger.Info("attempt to store dialog into nil db", zap.Uint64("user_id", userID), zap.Any("dialog", messages))
		return 0, nil
	}

	var id uint64
	if err := c.db.QueryRowx(storeDialogQuery, userID, messages).Scan(&id); err != nil {
		c.logger.Error("store dialog sql query process", zap.Uint64("user_id", userID), zap.Error(err))
		return 0, err
	}

	c.logger.Info("store dialog", zap.Uint64("user_id", userID), zap.Uint64("id", id))
	return id, nil
}

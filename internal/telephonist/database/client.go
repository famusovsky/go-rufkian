package database

import (
	"errors"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
)

type client struct {
	db        sqlx.Ext
	arenaPool *fastjson.ArenaPool
	logger    *zap.Logger
}

func NewClient(db sqlx.Ext, logger *zap.Logger) IClient {
	return client{
		db:        db,
		arenaPool: &fastjson.ArenaPool{},
		logger:    logger,
	}
}

func (c client) StoreDialog(dialog model.Dialog) (model.Dialog, error) {
	if c.db == nil {
		c.logger.Error("attempt to user empty db")
		return model.Dialog{}, errors.New("attempt to user empty db")
	}

	if len(dialog.Messages) == 0 {
		c.logger.Error("store empty dialog", zap.String("user_id", dialog.UserID))
		return model.Dialog{}, model.ErrEmptyDialog
	}

	arena := c.arenaPool.Get()
	arr := arena.NewArray()
	for i, msg := range dialog.Messages {
		obj := arena.NewObject()
		obj.Set("role", arena.NewString(string(msg.Role)))
		obj.Set("content", arena.NewString(msg.Content))
		if msg.Translation != nil {
			obj.Set("translation", arena.NewString(*msg.Translation))
		}
		arr.SetArrayItem(i, obj)
	}
	c.arenaPool.Put(arena)

	// XXX probably must convert messages to text, have to check
	var id string
	if err := c.db.QueryRowx(storeDialogQuery, dialog.UserID, dialog.StartTime, dialog.DurationS, arr.MarshalTo(nil)).Scan(&id); err != nil {
		c.logger.Error("store dialog sql query process", zap.String("user_id", dialog.UserID), zap.Error(err))
		return model.Dialog{}, err
	}

	dialog.ID = id

	c.logger.Info("store dialog", zap.String("user_id", dialog.UserID), zap.String("id", id))
	return dialog, nil
}

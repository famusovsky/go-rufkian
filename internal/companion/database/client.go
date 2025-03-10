package database

import (
	"errors"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fastjson"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// XXX probably should move logger from here

type client struct {
	db         sqlx.Ext
	parserPool *fastjson.ParserPool
	logger     *zap.Logger
}

func NewClient(db sqlx.Ext, logger *zap.Logger) IClient {
	return client{
		db:         db,
		parserPool: &fastjson.ParserPool{},
		logger:     logger,
	}
}

var errNilDB = errors.New("attempt to use nil db")

func (c client) AddUser(user model.User) (model.User, error) {
	if c.db == nil {
		c.logger.Error("attempt to store user into nil db")
		return model.User{}, errNilDB
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		c.logger.Error("generate hashed password", zap.Error(err))
		return model.User{}, err
	}

	user.Password = string(hashedPassword)

	if err := c.db.QueryRowx(addUserQuery, user.Email, user.Password).StructScan(&user); err != nil {
		c.logger.Error("store user sql query process", zap.Any("user", user), zap.Error(err))
		return model.User{}, err
	}

	c.logger.Info("store user", zap.Any("user", user))
	return user, nil
}

func (c client) GetUser(id string) (model.User, error) {
	if c.db == nil {
		c.logger.Warn("attempt to get user from nil db")
		return model.User{}, errNilDB
	}

	var user model.User
	if err := c.db.QueryRowx(getUserQuery, id).StructScan(&user); err != nil {
		c.logger.Error("get user sql query process", zap.Any("user", user), zap.Error(err))
		return model.User{}, err
	}

	c.logger.Info("get user", zap.Any("user", user))
	return user, nil
}
func (c client) GetUserByCredentials(user model.User) (model.User, error) {
	if c.db == nil {
		c.logger.Error("attempt to get user by creds from nil db")
		return model.User{}, errNilDB
	}

	inputPassword := user.Password

	if err := c.db.QueryRowx(getUserByCredsQuery, user.Email).StructScan(&user); err != nil {
		c.logger.Error("get user by creds sql query process", zap.Any("user_email", user.Email), zap.Error(err))
		return model.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputPassword)); err != nil {
		c.logger.Error("compare user password and hash", zap.Any("user_email", user.Email), zap.Error(err))
		return model.User{}, err
	}

	c.logger.Info("get user by creds", zap.Any("user", user))
	return user, nil
}

func (c client) GetDialog(id, userID string) (model.Dialog, error) {
	if c.db == nil {
		c.logger.Error("attempt to get dialog from nil db")
		return model.Dialog{}, nil
	}

	var dbDialog dbDialog
	if err := c.db.QueryRowx(getDialogQuery, id).StructScan(&dbDialog); err != nil {
		c.logger.Error("get gialog sql query process", zap.String("user_id", userID), zap.String("id", id))
		return model.Dialog{}, err
	}

	if dbDialog.UserID != userID {
		c.logger.Warn(
			"dialog userID and input userID does not match",
			zap.String("id", dbDialog.ID),
			zap.String("dialog user_id", dbDialog.UserID),
			zap.String("input user_id", userID),
		)
		return model.Dialog{}, errors.New("dialog userID and input userID does not match")
	}

	dialog, err := dbDialog.ToModel(c.parserPool)
	if err != nil {
		c.logger.Error("parse dialog messages", zap.String("user_id", userID), zap.String("id", id))
		return model.Dialog{}, err
	}

	c.logger.Info("get dialog", zap.String("user_id", userID), zap.String("id", id))
	return dialog, nil
}

func (c client) GetUserDialogs(userID string) (model.Dialogs, error) {
	if c.db == nil {
		c.logger.Error("attempt to get dialog from nil db")
		return nil, errNilDB
	}

	var dialogs model.Dialogs

	rows, err := c.db.Queryx(getUserDialogsQuery, userID)
	if err != nil {
		c.logger.Error("get user gialogs sql query process", zap.String("user_id", userID))
		return nil, err
	}

	for rows.Next() {
		var dbDialog dbDialog
		rows.StructScan(&dbDialog)
		dialog, err := dbDialog.ToModel(c.parserPool)
		if err != nil {
			c.logger.Error("parse dialog messages", zap.String("user_id", userID), zap.String("id", dbDialog.ID))
			continue
		}
		dialog.Messages = dialog.Messages[:1]
		dialogs = append(dialogs, dialog)
	}

	return dialogs, nil
}

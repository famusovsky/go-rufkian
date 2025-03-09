package database

import (
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

func (c client) AddUser(user model.User) (model.User, error) {
	if c.db == nil {
		c.logger.Warn("attempt to store user into nil db", zap.String("user_id", user.ID))
		return model.User{}, nil
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
		c.logger.Warn("attempt to get user from nil db", zap.String("user_id", id))
		return model.User{}, nil
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
		c.logger.Warn("attempt to get user by creds from nil db", zap.String("user_email", user.Email))
		return model.User{}, nil
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

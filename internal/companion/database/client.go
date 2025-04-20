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

// TODO do smth with errors: decide where you should log, and where just return err
func (c client) AddUser(user model.User) (model.User, error) {
	if c.db == nil {
		c.logger.Error("attempt to store user into nil db")
		return model.User{}, errNilDB
	}

	var id string
	if err := c.db.QueryRowx(addUserQuery, user.Email, user.Password).Scan(&id); err != nil {
		c.logger.Error("store user sql query process", zap.Any("user", user), zap.Error(err))
		return model.User{}, err
	}

	user.ID = id
	c.logger.Info("store user", zap.Any("user", user))
	return user, nil
}

func (c client) UpdateUser(user model.User) error {
	if c.db == nil {
		c.logger.Error("attempt to update user in nil db")
		return errNilDB
	}

	if _, err := c.db.Exec(updateUserQuery, user.ID, user.Email, user.Password, user.Key, user.TimeGoalM); err != nil {
		c.logger.Error("update user sql query process", zap.Any("user", user), zap.Error(err))
		return err
	}

	c.logger.Info("update user", zap.Any("user", user))
	return nil
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
		if err := rows.StructScan(&dbDialog); err != nil {
			c.logger.Error("scan dialog messages from sql row", zap.String("user_id", userID), zap.String("id", dbDialog.ID))
			continue
		}
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

func (c client) AddWordToUser(userID, word string) error {
	if c.db == nil {
		c.logger.Error("attempt to store word into nil db")
		return errNilDB
	}

	if _, err := c.db.Exec(addUserWordQuery, userID, word); err != nil {
		c.logger.Error("store user word sql query process", zap.String("user_id", userID), zap.String("word", word), zap.Error(err))
		return err
	}

	c.logger.Info("store user word sql query process", zap.String("user_id", userID), zap.String("word", word))
	return nil
}

func (c client) GetUserWords(userID string) ([]string, error) {
	if c.db == nil {
		c.logger.Warn("attempt to get user words from nil db")
		return nil, errNilDB
	}

	rows, err := c.db.Queryx(getUserWordsQuery, userID)
	if err != nil {
		c.logger.Error("get user words sql query process", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	var words []string
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			c.logger.Error("scan word from sql row", zap.String("user_id", userID), zap.Error(err))
			continue
		}
		words = append(words, word)
	}

	c.logger.Info("get user words", zap.Any("user_id", userID), zap.Any("words", words))
	return words, nil
}

func (c client) CheckUserWord(userID string, word string) (bool, error) {
	if c.db == nil {
		c.logger.Warn("attempt to check user word from nil db")
		return false, errNilDB
	}

	var relationExists bool
	if err := c.db.QueryRowx(checkUserWordQuery, userID, word).Scan(&relationExists); err != nil {
		c.logger.Error("check user word sql query process", zap.String("user_id", userID), zap.String("word", word))
		return false, err
	}

	return relationExists, nil
}

func (c client) DeleteWordFromUser(userID, word string) error {
	if c.db == nil {
		c.logger.Error("attempt to delete word from nil db")
		return errNilDB
	}

	if _, err := c.db.Exec(deleteWordFromUserQuery, userID, word); err != nil {
		c.logger.Error("delete user word sql query process", zap.String("user_id", userID), zap.String("word", word), zap.Error(err))
		return err
	}

	c.logger.Info("delete user word sql query process", zap.String("user_id", userID), zap.String("word", word))
	return nil
}

func (c client) AddWord(word, info, translation string) error {
	if c.db == nil {
		c.logger.Error("attempt to store word into nil db")
		return errNilDB
	}

	if _, err := c.db.Exec(addWordQuery, word, info, translation); err != nil {
		c.logger.Error("store word sql query process", zap.String("word", word), zap.Error(err))
		return err
	}

	c.logger.Info("store word sql query process", zap.String("word", word))
	return nil
}

func (c client) GetWordInfoAndTranslation(word string) (string, string, error) {
	if c.db == nil {
		c.logger.Error("attempt to get word from nil db")
		return "", "", errNilDB
	}

	var info, translation string
	if err := c.db.QueryRowx(getWordQuery, word).Scan(&info, &translation); err != nil {
		c.logger.Error("get word sql query process", zap.String("word", word))
		return "", "", err
	}

	c.logger.Info("get word sql query process", zap.String("word", word))
	return info, translation, nil
}

func (c client) GetDictionary(userID string) (model.Dictionary, error) {
	if c.db == nil {
		c.logger.Error("attempt to get dictionary from nil db")
		return model.Dictionary{}, errNilDB
	}

	var dictionary model.Dictionary
	if err := c.db.QueryRowx(getDictionaryQuery, userID).StructScan(&dictionary); err != nil {
		c.logger.Error("get dictionary sql query process", zap.String("user_id", userID))
		return model.Dictionary{}, err
	}

	c.logger.Info("get dictionary sql query process", zap.String("user_id", userID))
	return dictionary, nil
}

func (c client) CheckDictionaryNeedUpdate(userID, hash string) (bool, error) {
	if c.db == nil {
		c.logger.Error("attempt to get dictionary hash from nil db")
		return false, errNilDB
	}

	var dbHash string
	if err := c.db.QueryRowx(getDictionaryHashQuery, userID).Scan(&dbHash); err != nil {
		c.logger.Error("get dictionary hash sql query process", zap.String("user_id", userID))
		return false, err
	}

	return dbHash != hash, nil
}

func (c client) UpdateDictionary(dictionary model.Dictionary) error {
	if c.db == nil {
		c.logger.Error("attempt to update dictionary in nil db")
		return errNilDB
	}

	if _, err := c.db.Exec(updateDictionaryQuery, dictionary.UserID, dictionary.Hash, dictionary.Apkg); err != nil {
		c.logger.Error("update dictionary sql query process", zap.String("user_id", dictionary.UserID), zap.Error(err))
		return err
	}

	c.logger.Info("update dictionary sql query process", zap.String("user_id", dictionary.UserID))
	return nil
}

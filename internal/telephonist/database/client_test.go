package database

import (
	"database/sql"
	"testing"
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := NewClient(nil, logger)
	require.NotNil(t, client)
}

type minimalMockDB struct{}

func (m minimalMockDB) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return nil
}

func (m minimalMockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m minimalMockDB) BindNamed(query string, arg interface{}) (string, []interface{}, error) {
	return query, nil, nil
}

func (m minimalMockDB) DriverName() string {
	return "mock"
}

func (m minimalMockDB) Get(dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m minimalMockDB) MustExec(query string, args ...interface{}) sql.Result {
	return nil
}

func (m minimalMockDB) NamedExec(query string, arg interface{}) (sql.Result, error) {
	return nil, nil
}

func (m minimalMockDB) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return nil, nil
}

func (m minimalMockDB) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return nil, nil
}

func (m minimalMockDB) Preparex(query string) (*sqlx.Stmt, error) {
	return nil, nil
}

func (m minimalMockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (m minimalMockDB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return nil, nil
}

func (m minimalMockDB) Rebind(query string) string {
	return query
}

func (m minimalMockDB) Select(dest interface{}, query string, args ...interface{}) error {
	return nil
}

func TestStoreDialog(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		client := NewClient(nil, logger)

		dialog := model.Dialog{
			UserID:    "user123",
			StartTime: time.Now(),
			DurationS: 60,
			Messages: []model.Message{
				{
					Role:    model.UserRole,
					Content: "Hello",
				},
			},
		}

		_, err := client.StoreDialog(dialog)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty db")
	})

	t.Run("empty dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		mockDB := minimalMockDB{}
		client := NewClient(mockDB, logger)

		dialog := model.Dialog{
			UserID:    "user123",
			StartTime: time.Now(),
			DurationS: 60,
			Messages:  []model.Message{},
		}

		_, err := client.StoreDialog(dialog)

		require.Error(t, err)
		assert.Equal(t, model.ErrEmptyDialog, err)
	})
}

func TestUpdateDialog(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		client := NewClient(nil, logger)

		dialog := model.Dialog{
			ID:     "dialog123",
			UserID: "user123",
			Messages: []model.Message{
				{
					Role:    model.UserRole,
					Content: "Hello",
				},
			},
		}

		err := client.UpdateDialog(dialog)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty db")
	})

	t.Run("empty dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		mockDB := minimalMockDB{}
		client := NewClient(mockDB, logger)

		dialog := model.Dialog{
			ID:       "dialog123",
			UserID:   "user123",
			Messages: []model.Message{},
		}

		err := client.UpdateDialog(dialog)

		require.Error(t, err)
		assert.Equal(t, model.ErrEmptyDialog, err)
	})

	t.Run("dialog without ID", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		mockDB := minimalMockDB{}
		client := NewClient(mockDB, logger)

		dialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{
					Role:    model.UserRole,
					Content: "Hello",
				},
			},
		}

		err := client.UpdateDialog(dialog)

		require.Error(t, err)
		assert.Equal(t, model.ErrDialogWithoutID, err)
	})
}

package database

import (
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/valyala/fastjson"
)

type dbDialog struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Messages  string    `db:"messages"`
	StartTime time.Time `db:"start_time"`
	DurationS int       `db:"duration_s"`
}

func (dbDialog *dbDialog) ToModel(pool *fastjson.ParserPool) (model.Dialog, error) {
	parser := pool.Get()
	defer pool.Put(parser)

	rawMessages, err := parser.Parse(dbDialog.Messages)
	if err != nil {
		return model.Dialog{}, err
	}

	arr, err := rawMessages.Array()
	if err != nil {
		return model.Dialog{}, err
	}

	messages := make(model.Messages, 0, len(arr))
	for _, rawMessage := range arr {
		var msg model.Message
		msg.Role = model.Role(rawMessage.GetStringBytes("role"))
		msg.Content = string(rawMessage.GetStringBytes("content"))
		if translation := rawMessage.GetStringBytes("translation"); translation != nil {
			translationString := string(translation)
			msg.Translation = &translationString
		}
		messages = append(messages, msg)
	}

	return model.Dialog{
		ID:        dbDialog.ID,
		UserID:    dbDialog.UserID,
		Messages:  messages,
		StartTime: dbDialog.StartTime,
		DurationS: dbDialog.DurationS,
	}, nil
}

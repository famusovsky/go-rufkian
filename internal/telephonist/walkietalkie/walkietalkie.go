package walkietalkie

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/internal/telephonist/database"
	"github.com/famusovsky/go-rufkian/internal/telephonist/translator"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type controller struct {
	// FIXME situation when on the several apps same api key is used
	// TODO better to use some normal cache, Redis for example
	dialogs    sync.Map
	dbClient   database.IClient
	translator translator.IClient
	client     *resty.Client
	logger     *zap.Logger
}

func New(db sqlx.Ext, logger *zap.Logger, translator translator.IClient) IController {
	logger.Info("create walkie talkie controller")
	return &controller{
		dialogs:  sync.Map{},
		dbClient: database.NewClient(db, logger),
		client: resty.New().
			SetDisableWarn(true).
			SetAllowMethodGetPayload(true).
			SetContentLength(true).
			SetBaseURL(mistralChatCompletionsURL).
			SetTimeout(20 * time.Second),
		logger:     logger,
		translator: translator,
	}
}

func (c *controller) Talk(userID string, key, input string) string {
	c.logger.Info("talk request", zap.String("user_id", userID), zap.String("input", input))

	var dialog model.Dialog
	if historyRaw, ok := c.dialogs.Load(userID); ok {
		history, _ := historyRaw.(model.Dialog)
		// FIXME this log probably must be simplified -- could be very big
		c.logger.Info("dialog process history", zap.String("user_id", userID), zap.Any("history", history))
		dialog = history
	} else {
		c.logger.Info("create new dialog process", zap.String("user_id", userID))
		dialog = model.Dialog{
			Messages:  model.Messages{{Role: model.SystemRole, Content: chatSystemContent}},
			StartTime: time.Now(),
			UserID:    userID,
		}
	}
	dialog.Messages = append(dialog.Messages, model.Message{Role: model.UserRole, Content: input})
	dialog.UpdatedAt = time.Now()

	request := mistralRequest{
		Model:    model.MistralSmall,
		Messages: append(dialog.Messages, model.Message{Role: model.AssistantRole, Content: chatPrefixContent, Prefix: true}),
	}

	response, err := c.getMistralResponse(key, request)
	if err != nil {
		c.logger.Error("response from mistral", zap.String("user_id", userID), zap.Error(err))
		return ""
	}

	msg := withoutChatPrefix(response.Message())
	if msg.Empty() {
		c.logger.Warn("empty answer from mistral", zap.String("user_id", userID))
		return ""
	}
	dialog.Messages = append(dialog.Messages, msg)
	c.dialogs.Swap(userID, dialog)

	c.logger.Info("answer from mistral", zap.String("user_id", userID), zap.String("response", msg.Content))
	return msg.Content
}

func (c *controller) Stop(userID string) (string, error) {
	dialogRaw, ok := c.dialogs.LoadAndDelete(userID)
	if !ok {
		c.logger.Warn("delete dialog process", zap.String("user_id", userID), zap.Error(model.ErrNoHistoryFound))
		return "", nil
	}

	dialog, _ := dialogRaw.(model.Dialog)
	dialog.Messages = dialog.Messages.WithoutSystem()
	if len(dialog.Messages) < 2 {
		return "", nil
	}

	duration := time.Since(dialog.StartTime)
	dialog.DurationS = int(duration.Seconds())
	dialog, err := c.dbClient.StoreDialog(dialog)

	go func(dialog model.Dialog) {
		texts := make([]string, 0, len(dialog.Messages))
		for _, msg := range dialog.Messages {
			texts = append(texts, msg.Content)
		}

		translated, err := c.translator.Translate(texts)
		if err != nil {
			c.logger.Error("translating dialog", zap.String("dialog_id", dialog.ID), zap.Error(err))
			return
		}

		for i, t := range translated {
			dialog.Messages[i].Translation = &t
		}

		if err := c.dbClient.UpdateDialog(dialog); err != nil {
			c.logger.Error("store dialog translation", zap.String("dialog_id", dialog.ID), zap.Error(err))
		}
	}(dialog)

	return dialog.ID, err
}

func (c *controller) CleanUp() {
	now := time.Now()
	c.dialogs.Range(func(key, value any) bool {
		userID, keyOk := key.(string)
		dialog, valOk := value.(model.Dialog)
		if !keyOk || !valOk {
			return true
		}
		if dialog.UpdatedAt.Add(10 * time.Minute).Before(now) {
			c.Stop(userID)
		}
		return true
	})
}

type mistralRequest struct {
	Model    model.MistralModel `json:"model"`
	Messages model.Messages     `json:"messages"`
}

type mistralResponseFinishReason string

type mistralChoice struct {
	Index        int                         `json:"index"`
	Message      model.Message               `json:"message"`
	FinishReason mistralResponseFinishReason `json:"finish_reason"`
}

type mistralResponse struct {
	Choices []mistralChoice `json:"choices"`
}

func (mr mistralResponse) Message() model.Message {
	if len(mr.Choices) > 0 {
		return mr.Choices[len(mr.Choices)-1].Message
	}
	return model.Message{}
}

func (c *controller) getMistralResponse(key string, payload mistralRequest) (mistralResponse, error) {
	req := c.client.R().
		SetContentType("application/json").
		SetBody(payload).
		SetAuthToken(key).
		SetMethod(resty.MethodPost)

	response, err := req.Send()
	if err != nil {
		return mistralResponse{}, err
	}

	var ret mistralResponse
	if err := json.Unmarshal(response.Bytes(), &ret); err != nil {
		return mistralResponse{}, err
	}

	return ret, nil
}

func withoutChatPrefix(msg model.Message) model.Message {
	msg.Prefix = false
	if len(msg.Content) > 0 && msg.Role == model.AssistantRole {
		withoutPrefix := msg.Content[len(chatPrefixContent):]
		msg.Content = strings.TrimSpace(withoutPrefix)
	}
	return msg
}

package walkietalkie

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/internal/telephonist/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type controller struct {
	// FIXME situation when on the several apps same api key is used
	dialogs  sync.Map
	dbClient database.IClient
	client   *resty.Client
	logger   *zap.Logger
}

func New(db sqlx.Ext, logger *zap.Logger) IController {
	logger.Info("create walkie talkie controller")
	return &controller{
		dialogs:  sync.Map{},
		dbClient: database.NewClient(db, logger),
		client: resty.New().
			SetDisableWarn(true).
			SetAllowMethodGetPayload(true).
			SetContentLength(true).
			SetBaseURL(mistralChatCompletionsURL).
			SetTimeout(1 * time.Minute),
		logger: logger,
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

func (c *controller) Stop(userID, key string) (string, error) {
	dialogRaw, ok := c.dialogs.LoadAndDelete(userID)
	if !ok {
		c.logger.Warn("delete dialog process", zap.String("user_id", userID), zap.Error(model.ErrNoHistoryFound))
		return "", nil
	}

	dialog, _ := dialogRaw.(model.Dialog)
	dialog, err := c.dbClient.StoreDialog(dialog)

	go func() {
		dialog := dialog
		c.getTranslation(key, &dialog)
		dialog, err := c.dbClient.StoreDialog(dialog)
		if err != nil {
			c.logger.Error("store dialog translation", zap.String("dialog_id", dialog.ID), zap.Error(err))
		}
	}()

	return dialog.ID, err
}

type mistralRequest struct {
	Model    model.MistralModel `json:"model"`
	Messages model.Messages     `json:"messages"`
	// TODO stream
	// Stream   bool     `json:"stream"`
	// Stop     []string `json:"stop"`
}

type mistralResponceFinishReason string

type mistralChoice struct {
	Index        int                         `json:"index"`
	Message      model.Message               `json:"message"`
	FinishReason mistralResponceFinishReason `json:"finish_reason"`
}

type mistralResponce struct {
	Choices []mistralChoice `json:"choices"`
}

func (mr mistralResponce) Message() model.Message {
	if len(mr.Choices) > 0 {
		return mr.Choices[len(mr.Choices)-1].Message
	}
	return model.Message{}
}

func (c *controller) getTranslation(key string, dialog *model.Dialog) {
	wg := sync.WaitGroup{}
	wg.Add(len(dialog.Messages))
	for i, msg := range dialog.Messages {
		go func() {
			defer wg.Done()
			req := mistralRequest{
				Model: model.MistralSmall,
				Messages: model.Messages{
					model.Message{Role: model.SystemRole, Content: translationSystemContent},
					model.Message{Role: model.UserRole, Content: msg.Content},
					model.Message{Role: model.AssistantRole, Content: translationPrefixContent, Prefix: true},
				},
			}
			responce, err := c.getMistralResponse(key, req)
			if err != nil {
				c.logger.Error("translate message", zap.Error(err), zap.String("message", msg.Content))
				return
			}
			translation := withoutTranslationPrefix(responce.Message()).Content
			dialog.Messages[i].Translation = &translation
		}()
	}
	wg.Wait()
}

func (c *controller) getMistralResponse(key string, payload mistralRequest) (mistralResponce, error) {
	req := c.client.R().
		SetContentType("application/json").
		SetBody(payload).
		SetAuthToken(key).
		SetMethod(resty.MethodPost)

	response, err := req.Send()
	if err != nil {
		return mistralResponce{}, err
	}

	var ret mistralResponce
	if err := json.Unmarshal(response.Bytes(), &ret); err != nil {
		return mistralResponce{}, err
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

func withoutTranslationPrefix(msg model.Message) model.Message {
	msg.Prefix = false
	if len(msg.Content) > 0 && msg.Role == model.AssistantRole {
		withoutPrefix := msg.Content[len(translationPrefixContent):]
		msg.Content = strings.TrimSpace(withoutPrefix)
	}
	return msg
}

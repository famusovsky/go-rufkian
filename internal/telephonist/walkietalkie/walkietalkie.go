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
	// TODO use some id instead of api key -- it is vastly insecure
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

func (c *controller) Talk(key string, input string) string {
	c.logger.Info("talk request", zap.String("key", key), zap.String("input", input))

	var messages model.Messages
	if historyRaw, ok := c.dialogs.Load(key); ok {
		history, _ := historyRaw.(model.Messages)
		// FIXME this log probably must be simplified -- could be very big
		c.logger.Info("dialog process history", zap.String("key", key), zap.Any("history", history))
		messages = history
	} else {
		c.logger.Info("create new dialog process", zap.String("key", key))
		messages = model.Messages{{Role: model.SystemRole, Content: systemContent}}
	}
	messages = append(messages, model.Message{Role: model.UserRole, Content: input})

	request := mistralRequest{
		Model:    model.MistralSmall,
		Messages: append(messages, model.Message{Role: model.AssistantRole, Content: prefixContent, Prefix: true}),
	}

	response, err := c.getMistralResponse(key, request)
	if err != nil {
		c.logger.Error("response from mistral", zap.String("key", key), zap.Error(err))
		return ""
	}

	msg := withoutPrefix(response.Message())
	if msg.Empty() {
		c.logger.Warn("empty answer from mistral", zap.String("key", key))
		return ""
	}
	c.dialogs.Swap(key, append(messages, msg))

	c.logger.Info("answer from mistral", zap.String("key", key), zap.String("response", msg.Content))
	return msg.Content
}

func (c *controller) Stop(key string) (uint64, error) {
	messagesRaw, ok := c.dialogs.LoadAndDelete(key)
	if !ok {
		c.logger.Warn("delete dialog process", zap.String("key", key), zap.Error(model.ErrNoHistoryFound))
		return 0, nil
	}

	messages, _ := messagesRaw.(model.Messages)
	id, err := c.dbClient.StoreDialog(key, messages)

	return id, err
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

func withoutPrefix(msg model.Message) model.Message {
	msg.Prefix = false
	if len(msg.Content) > 0 && msg.Role == model.AssistantRole {
		withoutPrefix := msg.Content[len(prefixContent):]
		msg.Content = strings.TrimSpace(withoutPrefix)
	}
	return msg
}

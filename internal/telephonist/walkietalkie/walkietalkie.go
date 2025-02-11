package walkietalkie

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type controller struct {
	// TODO use some id instead of api key -- it is vastly insecure
	// FIXME situation when on the several apps same api key is used
	dialogs sync.Map
	client  *resty.Client
	logger  *zap.Logger
}

func New(logger *zap.Logger) IController {
	logger.Info("create walkie talkie controller")
	return &controller{
		dialogs: sync.Map{},
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

	msg := response.Message()
	if msg.Empty() {
		c.logger.Warn("empty answer from mistral", zap.String("key", key))
		return ""
	}
	c.dialogs.Swap(key, append(messages, msg))

	answer := withoutPrefix(msg)
	c.logger.Info("answer from mistral", zap.String("key", key), zap.String("response", answer))
	return answer
}

func (c *controller) Stop(key string) (string, error) {
	_, ok := c.dialogs.LoadAndDelete(key)
	var err error
	if !ok {
		err = model.ErrNoHistoryFound
	}

	// TODO store history in db, get and return db id
	c.logger.Info("delete dialog process", zap.String("key", key), zap.Error(err))
	return "", err
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

func withoutPrefix(msg model.Message) string {
	if len(msg.Content) > 0 && msg.Role == model.AssistantRole {
		withoutPrefix := msg.Content[len(prefixContent):]
		return strings.TrimSpace(withoutPrefix)
	}
	return ""
}

package walkietalkie

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/internal/telephonist/database"
	"github.com/famusovsky/go-rufkian/internal/telephonist/translator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type mockController struct {
	dialogs                sync.Map
	dbClient               database.IClient
	translator             translator.IClient
	logger                 *zap.Logger
	client                 interface{}
	getMistralResponseFunc func(key string, payload mistralRequest) (mistralResponse, error)
}

func (c *mockController) Talk(userID, key, input string) string {
	c.logger.Info("talk request", zap.String("user_id", userID), zap.String("input", input))

	var dialog model.Dialog
	if historyRaw, ok := c.dialogs.Load(userID); ok {
		history, _ := historyRaw.(model.Dialog)
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

	response, err := c.getMistralResponseFunc(key, request)
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
	c.dialogs.Store(userID, dialog)

	c.logger.Info("answer from mistral", zap.String("user_id", userID), zap.String("response", msg.Content))
	return msg.Content
}

func (c *mockController) Stop(userID string) (string, error) {
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

func (c *mockController) CleanUp() {
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

func TestNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	translatorMock := translator.NewClientMock(ctrl)

	controller := New(nil, logger, translatorMock)

	require.NotNil(t, controller)
}

func TestTalk(t *testing.T) {
	t.Run("new dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		getMistralResponseCalled := false
		getMistralResponseFunc := func(key string, payload mistralRequest) (mistralResponse, error) {
			getMistralResponseCalled = true
			assert.Equal(t, "key123", key)
			assert.Equal(t, model.MistralSmall, payload.Model)
			assert.Len(t, payload.Messages, 3)
			assert.Equal(t, model.SystemRole, payload.Messages[0].Role)
			assert.Equal(t, model.UserRole, payload.Messages[1].Role)
			assert.Equal(t, "hello", payload.Messages[1].Content)
			assert.Equal(t, model.AssistantRole, payload.Messages[2].Role)
			assert.Equal(t, chatPrefixContent, payload.Messages[2].Content)
			assert.True(t, payload.Messages[2].Prefix)

			return mistralResponse{
				Choices: []mistralChoice{
					{
						Message: model.Message{
							Role:    model.AssistantRole,
							Content: chatPrefixContent + "Hallo!",
						},
					},
				},
			}, nil
		}

		controller := &mockController{
			dialogs:                sync.Map{},
			dbClient:               dbMock,
			translator:             translatorMock,
			logger:                 logger,
			client:                 nil,
			getMistralResponseFunc: getMistralResponseFunc,
		}

		result := controller.Talk("user123", "key123", "hello")

		assert.True(t, getMistralResponseCalled)
		assert.Equal(t, "Hallo!", result)

		value, ok := controller.dialogs.Load("user123")
		assert.True(t, ok)
		dialog, ok := value.(model.Dialog)
		assert.True(t, ok)
		assert.Equal(t, "user123", dialog.UserID)
		assert.Len(t, dialog.Messages, 3)
		assert.Equal(t, model.SystemRole, dialog.Messages[0].Role)
		assert.Equal(t, model.UserRole, dialog.Messages[1].Role)
		assert.Equal(t, "hello", dialog.Messages[1].Content)
		assert.Equal(t, model.AssistantRole, dialog.Messages[2].Role)
		assert.Equal(t, "Hallo!", dialog.Messages[2].Content)
	})

	t.Run("existing dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		existingDialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{Role: model.SystemRole, Content: chatSystemContent},
				{Role: model.UserRole, Content: "hello"},
				{Role: model.AssistantRole, Content: "Hallo!"},
			},
			StartTime: time.Now(),
			UpdatedAt: time.Now(),
		}

		getMistralResponseCalled := false
		getMistralResponseFunc := func(key string, payload mistralRequest) (mistralResponse, error) {
			getMistralResponseCalled = true
			assert.Equal(t, "key123", key)
			assert.Equal(t, model.MistralSmall, payload.Model)

			var systemMsg, userHelloMsg, assistantHalloMsg, userHowAreYouMsg *model.Message

			for i := range payload.Messages {
				msg := &payload.Messages[i]
				if msg.Role == model.SystemRole {
					systemMsg = msg
				} else if msg.Role == model.UserRole && msg.Content == "hello" {
					userHelloMsg = msg
				} else if msg.Role == model.AssistantRole && msg.Content == "Hallo!" {
					assistantHalloMsg = msg
				} else if msg.Role == model.UserRole && msg.Content == "how are you?" {
					userHowAreYouMsg = msg
				}
			}

			assert.NotNil(t, systemMsg)
			assert.NotNil(t, userHelloMsg)
			assert.NotNil(t, assistantHalloMsg)
			assert.NotNil(t, userHowAreYouMsg)

			return mistralResponse{
				Choices: []mistralChoice{
					{
						Message: model.Message{
							Role:    model.AssistantRole,
							Content: chatPrefixContent + "Mir geht es gut, danke!",
						},
					},
				},
			}, nil
		}

		controller := &mockController{
			dialogs:                sync.Map{},
			dbClient:               dbMock,
			translator:             translatorMock,
			logger:                 logger,
			client:                 nil,
			getMistralResponseFunc: getMistralResponseFunc,
		}
		controller.dialogs.Store("user123", existingDialog)

		result := controller.Talk("user123", "key123", "how are you?")

		assert.True(t, getMistralResponseCalled)
		assert.Equal(t, "Mir geht es gut, danke!", result)

		value, ok := controller.dialogs.Load("user123")
		assert.True(t, ok)
		dialog, ok := value.(model.Dialog)
		assert.True(t, ok)
		assert.Equal(t, "user123", dialog.UserID)
		assert.Equal(t, "user123", dialog.UserID)

		var systemMsg, userHelloMsg, assistantHalloMsg, userHowAreYouMsg, assistantResponseMsg *model.Message

		for i := range dialog.Messages {
			msg := &dialog.Messages[i]
			if msg.Role == model.SystemRole {
				systemMsg = msg
			} else if msg.Role == model.UserRole && msg.Content == "hello" {
				userHelloMsg = msg
			} else if msg.Role == model.AssistantRole && msg.Content == "Hallo!" {
				assistantHalloMsg = msg
			} else if msg.Role == model.UserRole && msg.Content == "how are you?" {
				userHowAreYouMsg = msg
			} else if msg.Role == model.AssistantRole && msg.Content == "Mir geht es gut, danke!" {
				assistantResponseMsg = msg
			}
		}

		assert.NotNil(t, systemMsg)
		assert.NotNil(t, userHelloMsg)
		assert.NotNil(t, assistantHalloMsg)
		assert.NotNil(t, userHowAreYouMsg)
		assert.NotNil(t, assistantResponseMsg)
	})

	t.Run("mistral error", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		getMistralResponseCalled := false
		getMistralResponseFunc := func(key string, payload mistralRequest) (mistralResponse, error) {
			getMistralResponseCalled = true
			return mistralResponse{}, errors.New("mistral error")
		}

		controller := &mockController{
			dialogs:                sync.Map{},
			dbClient:               dbMock,
			translator:             translatorMock,
			logger:                 logger,
			client:                 nil,
			getMistralResponseFunc: getMistralResponseFunc,
		}

		result := controller.Talk("user123", "key123", "hello")

		assert.True(t, getMistralResponseCalled)
		assert.Equal(t, "", result)
	})

	t.Run("empty response", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		getMistralResponseCalled := false
		getMistralResponseFunc := func(key string, payload mistralRequest) (mistralResponse, error) {
			getMistralResponseCalled = true
			return mistralResponse{
				Choices: []mistralChoice{
					{
						Message: model.Message{
							Role:    model.AssistantRole,
							Content: chatPrefixContent,
						},
					},
				},
			}, nil
		}

		controller := &mockController{
			dialogs:                sync.Map{},
			dbClient:               dbMock,
			translator:             translatorMock,
			logger:                 logger,
			client:                 nil,
			getMistralResponseFunc: getMistralResponseFunc,
		}

		result := controller.Talk("user123", "key123", "hello")

		assert.True(t, getMistralResponseCalled)
		assert.Equal(t, "", result)
	})
}

func TestStop(t *testing.T) {
	t.Run("no dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		controller := &mockController{
			dialogs:    sync.Map{},
			dbClient:   dbMock,
			translator: translatorMock,
			logger:     logger,
			client:     nil,
		}

		id, err := controller.Stop("user123")

		assert.NoError(t, err)
		assert.Equal(t, "", id)
	})

	t.Run("dialog with less than 2 messages", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		controller := &mockController{
			dialogs:    sync.Map{},
			dbClient:   dbMock,
			translator: translatorMock,
			logger:     logger,
			client:     nil,
		}

		dialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{Role: model.SystemRole, Content: chatSystemContent},
			},
			StartTime: time.Now(),
			UpdatedAt: time.Now(),
		}
		controller.dialogs.Store("user123", dialog)

		id, err := controller.Stop("user123")

		assert.NoError(t, err)
		assert.Equal(t, "", id)

		_, ok := controller.dialogs.Load("user123")
		assert.False(t, ok)
	})

	t.Run("valid dialog", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		controller := &mockController{
			dialogs:    sync.Map{},
			dbClient:   dbMock,
			translator: translatorMock,
			logger:     logger,
			client:     nil,
		}

		dialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{Role: model.SystemRole, Content: chatSystemContent},
				{Role: model.UserRole, Content: "hello"},
				{Role: model.AssistantRole, Content: "Hallo!"},
			},
			StartTime: time.Now().Add(-time.Minute),
			UpdatedAt: time.Now(),
		}
		controller.dialogs.Store("user123", dialog)

		dbMock.EXPECT().StoreDialog(gomock.Any()).DoAndReturn(func(d model.Dialog) (model.Dialog, error) {
			assert.Equal(t, "user123", d.UserID)
			assert.Len(t, d.Messages, 2)
			assert.Equal(t, model.UserRole, d.Messages[0].Role)
			assert.Equal(t, "hello", d.Messages[0].Content)
			assert.Equal(t, model.AssistantRole, d.Messages[1].Role)
			assert.Equal(t, "Hallo!", d.Messages[1].Content)
			assert.GreaterOrEqual(t, d.DurationS, 60)

			d.ID = "dialog123"
			return d, nil
		})

		translatorMock.EXPECT().Translate([]string{"hello", "Hallo!"}).Return([]string{"привет", "Привет!"}, nil)

		dbMock.EXPECT().UpdateDialog(gomock.Any()).DoAndReturn(func(d model.Dialog) error {
			assert.Equal(t, "dialog123", d.ID)
			assert.Equal(t, "user123", d.UserID)
			assert.Len(t, d.Messages, 2)
			assert.Equal(t, model.UserRole, d.Messages[0].Role)
			assert.Equal(t, "hello", d.Messages[0].Content)
			assert.Equal(t, "привет", *d.Messages[0].Translation)
			assert.Equal(t, model.AssistantRole, d.Messages[1].Role)
			assert.Equal(t, "Hallo!", d.Messages[1].Content)
			assert.Equal(t, "Привет!", *d.Messages[1].Translation)

			return nil
		})

		id, err := controller.Stop("user123")

		assert.NoError(t, err)
		assert.Equal(t, "dialog123", id)

		_, ok := controller.dialogs.Load("user123")
		assert.False(t, ok)

		time.Sleep(100 * time.Millisecond)
	})

	t.Run("translation error", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		controller := &mockController{
			dialogs:    sync.Map{},
			dbClient:   dbMock,
			translator: translatorMock,
			logger:     logger,
			client:     nil,
		}

		dialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{Role: model.SystemRole, Content: chatSystemContent},
				{Role: model.UserRole, Content: "hello"},
				{Role: model.AssistantRole, Content: "Hallo!"},
			},
			StartTime: time.Now().Add(-time.Minute),
			UpdatedAt: time.Now(),
		}
		controller.dialogs.Store("user123", dialog)

		dbMock.EXPECT().StoreDialog(gomock.Any()).DoAndReturn(func(d model.Dialog) (model.Dialog, error) {
			d.ID = "dialog123"
			return d, nil
		})

		translatorMock.EXPECT().Translate([]string{"hello", "Hallo!"}).Return(nil, errors.New("translation error"))

		id, err := controller.Stop("user123")

		assert.NoError(t, err)
		assert.Equal(t, "dialog123", id)

		_, ok := controller.dialogs.Load("user123")
		assert.False(t, ok)

		time.Sleep(100 * time.Millisecond)
	})

	t.Run("update dialog error", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		translatorMock := translator.NewClientMock(ctrl)
		dbMock := database.NewClientMock(ctrl)

		controller := &mockController{
			dialogs:    sync.Map{},
			dbClient:   dbMock,
			translator: translatorMock,
			logger:     logger,
			client:     nil,
		}

		dialog := model.Dialog{
			UserID: "user123",
			Messages: []model.Message{
				{Role: model.SystemRole, Content: chatSystemContent},
				{Role: model.UserRole, Content: "hello"},
				{Role: model.AssistantRole, Content: "Hallo!"},
			},
			StartTime: time.Now().Add(-time.Minute),
			UpdatedAt: time.Now(),
		}
		controller.dialogs.Store("user123", dialog)

		dbMock.EXPECT().StoreDialog(gomock.Any()).DoAndReturn(func(d model.Dialog) (model.Dialog, error) {
			d.ID = "dialog123"
			return d, nil
		})

		translatorMock.EXPECT().Translate([]string{"hello", "Hallo!"}).Return([]string{"привет", "Привет!"}, nil)

		dbMock.EXPECT().UpdateDialog(gomock.Any()).Return(errors.New("update error"))

		id, err := controller.Stop("user123")

		assert.NoError(t, err)
		assert.Equal(t, "dialog123", id)

		_, ok := controller.dialogs.Load("user123")
		assert.False(t, ok)

		time.Sleep(100 * time.Millisecond)
	})
}

func TestCleanUp(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	translatorMock := translator.NewClientMock(ctrl)
	dbMock := database.NewClientMock(ctrl)

	controller := &mockController{
		dialogs:    sync.Map{},
		dbClient:   dbMock,
		translator: translatorMock,
		logger:     logger,
		client:     nil,
	}

	oldDialog := model.Dialog{
		UserID: "user1",
		Messages: []model.Message{
			{Role: model.SystemRole, Content: chatSystemContent},
			{Role: model.UserRole, Content: "hello"},
			{Role: model.AssistantRole, Content: "Hallo!"},
		},
		StartTime: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-11 * time.Minute),
	}
	controller.dialogs.Store("user1", oldDialog)

	newDialog := model.Dialog{
		UserID: "user2",
		Messages: []model.Message{
			{Role: model.SystemRole, Content: chatSystemContent},
			{Role: model.UserRole, Content: "hello"},
			{Role: model.AssistantRole, Content: "Hallo!"},
		},
		StartTime: time.Now().Add(-time.Minute),
		UpdatedAt: time.Now(),
	}
	controller.dialogs.Store("user2", newDialog)

	dbMock.EXPECT().StoreDialog(gomock.Any()).DoAndReturn(func(d model.Dialog) (model.Dialog, error) {
		assert.Equal(t, "user1", d.UserID)
		d.ID = "dialog1"
		return d, nil
	})

	translatorMock.EXPECT().Translate(gomock.Any()).Return([]string{"привет", "Привет!"}, nil)

	dbMock.EXPECT().UpdateDialog(gomock.Any()).Return(nil)

	controller.CleanUp()

	time.Sleep(100 * time.Millisecond)

	_, ok1 := controller.dialogs.Load("user1")
	assert.False(t, ok1)

	_, ok2 := controller.dialogs.Load("user2")
	assert.True(t, ok2)
}

func TestWithoutChatPrefix(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		msg := model.Message{
			Role:    model.AssistantRole,
			Content: chatPrefixContent + "Hallo!",
			Prefix:  true,
		}

		result := withoutChatPrefix(msg)

		assert.Equal(t, model.AssistantRole, result.Role)
		assert.Equal(t, "Hallo!", result.Content)
		assert.False(t, result.Prefix)
	})

	t.Run("without prefix", func(t *testing.T) {
		msg := model.Message{
			Role:    model.UserRole,
			Content: "hello",
			Prefix:  false,
		}

		result := withoutChatPrefix(msg)

		assert.Equal(t, model.UserRole, result.Role)
		assert.Equal(t, "hello", result.Content)
		assert.False(t, result.Prefix)
	})
}

func TestMistralResponseMessage(t *testing.T) {
	t.Run("with choices", func(t *testing.T) {
		response := mistralResponse{
			Choices: []mistralChoice{
				{
					Message: model.Message{
						Role:    model.UserRole,
						Content: "hello",
					},
				},
				{
					Message: model.Message{
						Role:    model.AssistantRole,
						Content: "Hallo!",
					},
				},
			},
		}

		msg := response.Message()

		assert.Equal(t, model.AssistantRole, msg.Role)
		assert.Equal(t, "Hallo!", msg.Content)
	})

	t.Run("without choices", func(t *testing.T) {
		response := mistralResponse{
			Choices: []mistralChoice{},
		}

		msg := response.Message()

		assert.Equal(t, model.Role(""), msg.Role)
		assert.Equal(t, "", msg.Content)
	})
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/pkg/cookie"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestUserFromCtx(t *testing.T) {
	t.Run("returns user when present in context", func(t *testing.T) {
		app := fiber.New()
		expectedUser := model.User{ID: "user123"}

		app.Use(func(c *fiber.Ctx) error {
			ctx := context.WithValue(context.Background(), userValueKey, expectedUser)
			c.SetUserContext(ctx)
			return c.Next()
		})

		var retrievedUser model.User
		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			retrievedUser, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, expectedUser, retrievedUser)
	})

	t.Run("returns false when user not in context", func(t *testing.T) {
		app := fiber.New()

		var retrievedUser model.User
		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			retrievedUser, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, retrievedUser.ID)
	})
}

func TestSetContext(t *testing.T) {
	t.Run("sets user in context when cookie and DB lookup succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)
		logger := zap.NewNop()

		app := fiber.New()

		expectedUser := model.User{ID: "user123"}
		cookieValues := map[string]string{model.UserKey: "user123"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)
		mockDBClient.EXPECT().GetUser("user123").Return(expectedUser, nil)

		app.Use(SetContext(mockCookieHandler, mockDBClient, logger))

		var retrievedUser model.User
		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			retrievedUser, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, expectedUser, retrievedUser)
	})

	t.Run("does not set user when cookie read fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)
		logger := zap.NewNop()

		app := fiber.New()

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(nil, errors.New("cookie read error"))

		app.Use(SetContext(mockCookieHandler, mockDBClient, logger))

		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			_, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("does not set user when cookie has no user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)
		logger := zap.NewNop()

		app := fiber.New()

		cookieValues := map[string]string{"some_other_key": "value"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)

		app.Use(SetContext(mockCookieHandler, mockDBClient, logger))

		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			_, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("does not set user when DB lookup fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)
		logger := zap.NewNop()

		app := fiber.New()

		cookieValues := map[string]string{model.UserKey: "user123"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)
		mockDBClient.EXPECT().GetUser("user123").Return(model.User{}, errors.New("db error"))

		app.Use(SetContext(mockCookieHandler, mockDBClient, logger))

		var ok bool
		app.Get("/test", func(c *fiber.Ctx) error {
			_, ok = UserFromCtx(c)
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)

		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestCheckUser(t *testing.T) {
	t.Run("allows request when user is in context", func(t *testing.T) {
		app := fiber.New()
		expectedUser := model.User{ID: "user123"}

		app.Use(func(c *fiber.Ctx) error {
			ctx := context.WithValue(context.Background(), userValueKey, expectedUser)
			c.SetUserContext(ctx)
			return c.Next()
		})

		app.Use(CheckUser)

		handlerCalled := false
		app.Get("/test", func(c *fiber.Ctx) error {
			handlerCalled = true
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, handlerCalled)
	})

	t.Run("redirects to auth when user is not in context", func(t *testing.T) {
		app := fiber.New()

		app.Use(CheckUser)

		handlerCalled := false
		app.Get("/test", func(c *fiber.Ctx) error {
			handlerCalled = true
			return c.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode) // 302 Found (redirect)
		assert.Equal(t, "/auth", resp.Header.Get("Location"))
		assert.Equal(t, "body", resp.Header.Get("HX-Retarget"))
		assert.False(t, handlerCalled)
	})
}

func TestGetUser(t *testing.T) {
	t.Run("returns user when cookie and DB lookup succeed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)

		expectedUser := model.User{ID: "user123"}
		cookieValues := map[string]string{model.UserKey: "user123"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)
		mockDBClient.EXPECT().GetUser("user123").Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/", nil)
		user, err := getUser(req, mockCookieHandler, mockDBClient)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("returns error when cookie read fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(nil, errors.New("cookie read error"))

		req := httptest.NewRequest("GET", "/", nil)
		_, err := getUser(req, mockCookieHandler, mockDBClient)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "read cookie")
	})

	t.Run("returns error when cookie has no user ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)

		cookieValues := map[string]string{"some_other_key": "value"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)

		req := httptest.NewRequest("GET", "/", nil)
		_, err := getUser(req, mockCookieHandler, mockDBClient)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cookie has not user_id")
	})

	t.Run("returns error when DB lookup fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCookieHandler := cookie.NewHandlerMock(ctrl)
		mockDBClient := database.NewClientMock(ctrl)

		cookieValues := map[string]string{model.UserKey: "user123"}

		mockCookieHandler.EXPECT().Read(gomock.Any()).Return(cookieValues, nil)
		mockDBClient.EXPECT().GetUser("user123").Return(model.User{}, errors.New("db error"))

		req := httptest.NewRequest("GET", "/", nil)
		_, err := getUser(req, mockCookieHandler, mockDBClient)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get user from db")
	})
}

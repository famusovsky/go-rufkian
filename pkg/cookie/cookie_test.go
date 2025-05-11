package cookie

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSet(t *testing.T) {
	t.Run("sets cookie with key-value pairs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAdder := NewadderMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)
		expireDate := time.Now().Add(24 * time.Hour)

		mockAdder.EXPECT().Add("Set-Cookie", gomock.Any()).Times(1)

		err := handler.Set(mockAdder, expireDate, "key1", "value1", "key2", "value2")
		assert.NoError(t, err)
	})

	t.Run("returns nil when no key-value pairs are provided", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAdder := NewadderMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)
		expireDate := time.Now().Add(24 * time.Hour)

		err := handler.Set(mockAdder, expireDate)
		assert.NoError(t, err)
	})

	t.Run("returns nil when only one value is provided", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAdder := NewadderMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)
		expireDate := time.Now().Add(24 * time.Hour)

		err := handler.Set(mockAdder, expireDate, "key1")
		assert.NoError(t, err)
	})
}

func TestRead(t *testing.T) {
	t.Run("reads and decodes cookie values", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRequest := NewrequestMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)

		cookie := &http.Cookie{
			Name:  cookieName,
			Value: "encoded-value", // This would normally be an encoded value
		}

		mockRequest.EXPECT().Cookie(cookieName).Return(cookie, nil).Times(1)

		values, err := handler.Read(mockRequest)
		assert.Error(t, err)
		assert.Nil(t, values)
	})

	t.Run("returns error when cookie is not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRequest := NewrequestMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)

		mockRequest.EXPECT().Cookie(cookieName).Return(nil, errors.New("cookie not found")).Times(1)

		values, err := handler.Read(mockRequest)
		assert.Error(t, err)
		assert.Nil(t, values)
		assert.Contains(t, err.Error(), "read cookie from http request")
	})
}

func TestClear(t *testing.T) {
	t.Run("clears cookie", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAdder := NewadderMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)

		mockAdder.EXPECT().Add("Set-Cookie", gomock.Any()).Times(1)

		handler.Clear(mockAdder)
	})
}

func TestSetCookie(t *testing.T) {
	t.Run("adds cookie to adder", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAdder := NewadderMock(ctrl)
		cookieName := "test-cookie"
		handler := NewHandler(cookieName)
		expireDate := time.Now().Add(24 * time.Hour)

		mockAdder.EXPECT().Add("Set-Cookie", gomock.Any()).Times(1)

		_ = handler.Set(mockAdder, expireDate, "key1", "value1")
	})
}

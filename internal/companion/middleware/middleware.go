package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/pkg/cookie"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"go.uber.org/zap"
)

type authMiddlewareKey int

const (
	userValueKey authMiddlewareKey = iota
)

func UserFromCtx(c *fiber.Ctx) (model.User, bool) {
	ctx := c.UserContext()
	rawUser := ctx.Value(userValueKey)
	user, ok := rawUser.(model.User)
	return user, ok
}

func SetContext(
	cookieHandler cookie.IHandler,
	dbClient database.IClient,
	logger *zap.Logger,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req, err := adaptor.ConvertRequest(c, false)
		if err != nil {
			return fmt.Errorf("convert fiber context to http request: %w", err)
		}

		ctx := context.Background()

		if user, err := getUser(req, cookieHandler, dbClient); err == nil {
			ctx = context.WithValue(ctx, userValueKey, user)
		} else {
			logger.Warn("get user from cookie", zap.Error(err))
		}

		c.SetUserContext(ctx)

		return c.Next()
	}
}

func CheckUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, ok := UserFromCtx(c)
		if !ok {
			c.Set("HX-Retarget", "body")
			return c.Redirect("/auth")
		}
		return c.Next()
	}
}

func getUser(r *http.Request, cookieHandler cookie.IHandler, dbClient database.IClient) (model.User, error) {
	valueMap, err := cookieHandler.Read(r)
	if err != nil {
		return model.User{}, fmt.Errorf("read cookie: %w", err)
	}

	userID, ok := valueMap[model.UserKey]
	if !ok {
		return model.User{}, fmt.Errorf("cookie has not user_id")
	}

	user, err := dbClient.GetUser(userID)
	if err != nil {
		return model.User{}, fmt.Errorf("get user from db: %w", err)
	}

	return user, nil
}

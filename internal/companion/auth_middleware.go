package companion

import (
	"context"

	"github.com/famusovsky/go-rufkian/internal/model"
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

func (s *server) checkReg(c *fiber.Ctx) error {
	user, ok := s.getUser(c)
	if !ok {
		return c.Redirect("/auth")
	}

	c.SetUserContext(
		context.WithValue(
			context.Background(),
			userValueKey,
			user,
		),
	)

	return c.Next()
}

func (s *server) getUser(c *fiber.Ctx) (model.User, bool) {
	req, err := adaptor.ConvertRequest(c, false)
	if err != nil {
		s.logger.Error("convert fiber context to http request", zap.Error(err))
		return model.User{}, false
	}
	valueMap, err := s.cookieHandler.Read(req)
	if err != nil {
		s.logger.Warn("read cookie", zap.Error(err))
		return model.User{}, false
	}

	userID, ok := valueMap[userIDKey]
	if !ok {
		s.logger.Warn("cookie has not user_id")
		return model.User{}, false
	}

	user, err := s.dbClient.GetUser(userID)
	if err != nil {
		s.logger.Error("get user from db", zap.Error(err))
		return model.User{}, false
	}

	return user, true
}

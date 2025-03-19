package user

import (
	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type handlers struct {
	dbClient database.IClient
	logger   *zap.Logger
}

func NewHandlers(
	dbClient database.IClient,
	logger *zap.Logger,
) IHandlers {
	return &handlers{
		dbClient: dbClient,
		logger:   logger,
	}
}

func (h *handlers) GetInfo(c *fiber.Ctx) error {
	user, ok := middleware.UserFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.JSON(user)
}

type updateRequest struct {
	Email, Password, Key string
	Goal                 *int
}

func (h *handlers) Update(c *fiber.Ctx) error {
	user, ok := middleware.UserFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var updateRequest updateRequest
	h.logger.Info("got", zap.ByteString("body", c.Body()))
	if err := c.BodyParser(&updateRequest); err != nil {
		return render.ErrToResult(c, fiber.ErrBadRequest)
	}

	h.logger.Info("update", zap.Any("user", updateRequest), zap.String("pswd", updateRequest.Password))

	if len(updateRequest.Email) > 0 {
		user.Email = updateRequest.Email
	}
	if len(updateRequest.Password) >= 8 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateRequest.Password), 8)
		if err != nil {
			h.logger.Error("generate hashed password", zap.Error(err))
		} else {
			user.Password = string(hashedPassword)
		}
	}
	if len(updateRequest.Key) > 0 {
		user.Key = updateRequest.Key
	}
	if updateRequest.Goal != nil {
		user.TimeGoalM = *updateRequest.Goal
	}
	h.logger.Info("user", zap.Any("info", user), zap.String("pswd", user.Password))

	if err := h.dbClient.UpdateUser(user); err != nil {
		return render.ErrToResult(c, err)
	}

	c.Set("HX-Refresh", "true")
	return c.SendString("OK")
}

func (h *handlers) SettingsPage(c *fiber.Ctx) error {
	user, _ := middleware.UserFromCtx(c)
	return c.Render("settings", fiber.Map{
		"email": user.Email,
		"goal":  user.TimeGoalM,
		"key":   user.Key,
	}, "layouts/base")
}

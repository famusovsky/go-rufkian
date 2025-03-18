package user

import (
	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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
}

// TODO do better
func (h *handlers) Update(c *fiber.Ctx) error {
	user, ok := middleware.UserFromCtx(c)
	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var updateRequest updateRequest
	if err := c.BodyParser(&updateRequest); err != nil {
		return render.ErrToResult(c, fiber.ErrBadRequest)
	}

	if len(updateRequest.Email) > 0 {
		user.Email = updateRequest.Email
	}
	if len(updateRequest.Password) > 0 {
		user.Password = updateRequest.Password
	}
	if len(updateRequest.Key) > 0 {
		user.Key = &updateRequest.Key
	}

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
	}, "layouts/base")
}

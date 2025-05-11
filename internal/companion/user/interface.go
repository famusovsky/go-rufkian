package user

import (
	"github.com/gofiber/fiber/v2"
)

//go:generate mockgen -package user -mock_names IHandlers=HandlersMock -source ./interface.go -typed -destination interface.mock.gen.go
type IHandlers interface {
	GetInfo(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	SettingsPage(c *fiber.Ctx) error
}

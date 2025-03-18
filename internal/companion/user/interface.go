package user

import (
	"github.com/gofiber/fiber/v2"
)

type IHandlers interface {
	GetInfo(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	SettingsPage(c *fiber.Ctx) error
}

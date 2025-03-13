package auth

import (
	"github.com/gofiber/fiber/v2"
)

type IHandlers interface {
	SignUp(c *fiber.Ctx) error
	SignIn(c *fiber.Ctx) error
	SignOut(c *fiber.Ctx) error

	AuthPage(c *fiber.Ctx) error
}

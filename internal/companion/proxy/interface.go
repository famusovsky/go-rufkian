package proxy

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	Woerter(c *fiber.Ctx) error
}

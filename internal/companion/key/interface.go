package key

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	InsertPage(c *fiber.Ctx) error
	InstructionPage(c *fiber.Ctx) error
}

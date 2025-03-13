package dictionary

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	DictionaryPage(c *fiber.Ctx) error
	WordPage(c *fiber.Ctx) error
}

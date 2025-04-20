package dictionary

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	AddWord(c *fiber.Ctx) error
	DeleteWord(c *fiber.Ctx) error

	DictionaryPage(c *fiber.Ctx) error
	WordPage(c *fiber.Ctx) error

	GetApkg(c *fiber.Ctx) error
	ApkgInstructionPage(c *fiber.Ctx) error
}

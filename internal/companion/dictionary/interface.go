package dictionary

import "github.com/gofiber/fiber/v2"

//go:generate mockgen -package dictionary -mock_names IHandlers=HandlersMock -source ./interface.go -typed -destination interface.mock.gen.go
type IHandlers interface {
	AddWord(c *fiber.Ctx) error
	DeleteWord(c *fiber.Ctx) error

	DictionaryPage(c *fiber.Ctx) error
	WordPage(c *fiber.Ctx) error

	GetApkg(c *fiber.Ctx) error
	ApkgInstructionPage(c *fiber.Ctx) error
}

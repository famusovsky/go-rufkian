package key

import "github.com/gofiber/fiber/v2"

//go:generate mockgen -package key -mock_names IHandlers=HandlersMock -source ./interface.go -typed -destination interface.mock.gen.go
type IHandlers interface {
	InsertPage(c *fiber.Ctx) error
	InstructionPage(c *fiber.Ctx) error
}

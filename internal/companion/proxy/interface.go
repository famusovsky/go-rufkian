package proxy

import "github.com/gofiber/fiber/v2"

//go:generate mockgen -package proxy -mock_names IHandlers=HandlersMock -source ./interface.go -typed -destination interface.mock.gen.go
type IHandlers interface {
	Woerter(c *fiber.Ctx) error
}

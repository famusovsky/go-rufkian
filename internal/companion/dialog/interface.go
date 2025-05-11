package dialog

import "github.com/gofiber/fiber/v2"

//go:generate mockgen -package dialog -mock_names IHandlers=HandlersMock -source ./interface.go -typed -destination interface.mock.gen.go
type IHandlers interface {
	DialogPage(c *fiber.Ctx) error
	HistoryPage(c *fiber.Ctx) error
}

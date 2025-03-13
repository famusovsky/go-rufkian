package dialog

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	DialogPage(c *fiber.Ctx) error
	HistoryPage(c *fiber.Ctx) error
}

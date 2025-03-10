package dialog

import "github.com/gofiber/fiber/v2"

type IHandlers interface {
	RenderPage(c *fiber.Ctx) error
	RenderHistoryPage(c *fiber.Ctx) error
}

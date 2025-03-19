package key

import "github.com/gofiber/fiber/v2"

type handlers struct {
}

func NewHandlers() IHandlers {
	return &handlers{}
}

func (h *handlers) InsertPage(c *fiber.Ctx) error {
	return c.Render("insertKey", fiber.Map{}, "layouts/base")
}
func (h *handlers) InstructionPage(c *fiber.Ctx) error {
	return c.Render("instructionKey", fiber.Map{}, "layouts/base")
}

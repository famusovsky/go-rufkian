package companion

import "github.com/gofiber/fiber/v2"

// AUTH

func (s *server) RenderAuthPage(c *fiber.Ctx) error {
	return c.Render("auth", fiber.Map{}, "layouts/mini")
}

// BASE

func (s *server) RenderMainPage(c *fiber.Ctx) error {
	user, _ := UserFromCtx(c)
	return c.Render("main", fiber.Map{"email": user.Email}, "layouts/base")
}

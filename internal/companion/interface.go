package companion

import "github.com/gofiber/fiber/v2"

type IServer interface {
	Run()
	Shutdown()
}

type authHandlers interface {
	SignUp(c *fiber.Ctx) error
	SignIn(c *fiber.Ctx) error
	SignOut(c *fiber.Ctx) error

	RenderPage(c *fiber.Ctx) error
}

type dialogHandlers interface {
	RenderPage(c *fiber.Ctx) error
	RenderHistoryPage(c *fiber.Ctx) error
}

type proxyHandlers interface {
	Woerter(c *fiber.Ctx) error
	// ...
}

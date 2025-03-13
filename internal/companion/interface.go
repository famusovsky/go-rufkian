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

	AuthPage(c *fiber.Ctx) error
}

type dialogHandlers interface {
	DialogPage(c *fiber.Ctx) error
	HistoryPage(c *fiber.Ctx) error
}

type dictionaryHandlers interface {
	DictionaryPage(c *fiber.Ctx) error
	WordPage(c *fiber.Ctx) error
}

type proxyHandlers interface {
	Woerter(c *fiber.Ctx) error
	// ...
}

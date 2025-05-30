package companion

import (
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/gofiber/fiber/v2"
)

func (s *server) initRouter() {
	s.app.Static("/static", "./ui/static")
	s.app.Get("favicon.ico", func(c *fiber.Ctx) error {
		return c.SendFile("ui/static/favicon.ico")
	})

	s.app.Get("/call", func(c *fiber.Ctx) error {
		return c.SendString("call")
	})

	main := s.app.Group("/", middleware.CheckUserAgent)

	// TODO redirect from auth if user is already authed
	auth := main.Group("/auth")
	auth.Get("/", s.authHandlers.AuthPage)
	auth.Put("/", s.authHandlers.SignIn)
	auth.Post("/", s.authHandlers.SignUp)
	auth.Delete("/", s.authHandlers.SignOut)

	proxy := main.Group("/proxy")
	proxy.Get("/woerter/:q<string>", s.proxyHandlers.Woerter)

	withContext := main.Group("/", middleware.SetContext(s.cookieHandler, s.dbClient, s.logger))
	withUser := withContext.Group("/", middleware.CheckUser)

	key := withUser.Group("/key")
	key.Get("/insert", s.keyHandlers.InsertPage)
	key.Get("/instruction", s.keyHandlers.InstructionPage)

	user := withUser.Group("/user")
	user.Put("/", s.userHandlers.Update)
	user.Get("/", s.userHandlers.GetInfo)
	user.Get("/settings", s.userHandlers.SettingsPage)
	// TODO main page

	withUser.Get("/", s.dialogHandlers.HistoryPage)

	dialog := withUser.Group("/dialog")
	dialog.Get("/", s.dialogHandlers.HistoryPage)
	dialog.Get("/:id<int>", s.dialogHandlers.DialogPage)

	dictionary := withUser.Group("/dictionary")
	dictionary.Get("/", s.dictionaryHandlers.DictionaryPage)
	dictionary.Get("/apkg", s.dictionaryHandlers.GetApkg)
	dictionary.Get("/apkg/instruction", s.dictionaryHandlers.ApkgInstructionPage)
	// TODO save word in context (or make add/delete better)
	word := dictionary.Group("/:word<string>")
	word.Get("/", s.dictionaryHandlers.WordPage)
	word.Post("/", s.dictionaryHandlers.AddWord)
	word.Delete("/", s.dictionaryHandlers.DeleteWord)
}

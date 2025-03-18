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

	// TODO redirect from auth if user is already authed
	auth := s.app.Group("/auth")
	auth.Get("/", s.authHandlers.AuthPage)
	auth.Put("/", s.authHandlers.SignIn)
	auth.Post("/", s.authHandlers.SignUp)
	auth.Delete("/", s.authHandlers.SignOut)

	proxy := s.app.Group("/proxy")
	proxy.Get("/woerter/:q<string>", s.proxyHandlers.Woerter)

	withContext := s.app.Group("/", middleware.SetContext(s.cookieHandler, s.dbClient, s.logger))
	withUser := withContext.Group("/", middleware.CheckUser())
	withUser.Get("/auth/user", s.authHandlers.UserInfo)
	// TODO main page

	withUser.Get("/", s.dialogHandlers.HistoryPage)

	dialog := withUser.Group("/dialog")
	dialog.Get("/", s.dialogHandlers.HistoryPage)
	dialog.Get("/:id<int>", s.dialogHandlers.DialogPage)

	dictionary := withUser.Group("/dictionary")
	dictionary.Get("/", s.dictionaryHandlers.DictionaryPage)
	// TODO save word in context (or make add/delete better)
	word := dictionary.Group("/:word<string>")
	word.Get("/", s.dictionaryHandlers.WordPage)
	word.Post("/", s.dictionaryHandlers.AddWord, s.dictionaryHandlers.WordPage)
	word.Delete("/", s.dictionaryHandlers.DeleteWord, s.dictionaryHandlers.WordPage)
}

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

	withUser.Get("/", s.dialogHandlers.HistoryPage)
	withUser.Get("/dialog/:id<int>", s.dialogHandlers.DialogPage)
	withUser.Get("/dictionary", s.dictionaryHandlers.DictionaryPage)
	withUser.Get("/dictionary/word/:word<string>", s.dictionaryHandlers.WordPage)

	// TODO implement all other routes
}

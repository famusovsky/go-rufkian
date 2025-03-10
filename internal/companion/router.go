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

	auth := s.app.Group("/auth")
	auth.Get("/", s.authHandlers.RenderPage)
	auth.Put("/", s.authHandlers.SignIn)
	auth.Post("/", s.authHandlers.SignUp)
	auth.Delete("/", s.authHandlers.SignOut)

	withContext := s.app.Group("/", middleware.SetContext(s.cookieHandler, s.dbClient, s.logger))
	withUser := withContext.Group("/", middleware.CheckUser())

	withUser.Get("/", s.dialogHandlers.RenderHistoryPage)
	withUser.Get("/dialog/:id<int>", s.dialogHandlers.RenderPage)

	// TODO implement all other routes
}

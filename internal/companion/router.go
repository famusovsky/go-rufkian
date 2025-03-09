package companion

func (s *server) initRouter() {
	s.app.Static("/static", "./ui/static")

	auth := s.app.Group("/auth")
	auth.Get("/", s.RenderAuthPage)
	auth.Put("/", s.SignIn)
	auth.Post("/", s.SignUp)
	auth.Delete("/", s.signOut)

	base := s.app.Group("/", s.checkReg)
	base.Get("/", s.RenderMainPage)
	// TODO implement all other routes

	s.app.Get("/hello", s.hello)
	s.app.Get("favicon.ico", s.favicon)
}

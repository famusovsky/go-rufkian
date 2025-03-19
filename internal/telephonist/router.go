package telephonist

func (s *server) initRouter() {
	s.app.Post("/", s.Post)
	s.app.Delete("/", s.Delete)
}

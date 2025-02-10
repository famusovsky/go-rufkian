package telephonist

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				logger.Error(
					"route handle",
					zap.String("url", c.OriginalURL()),
					zap.String("method", c.Method()),
					zap.Error(err),
				)
				// TODO more appropriate error
				return c.SendStatus(fiber.StatusInternalServerError)
			},
		}),
		logger: logger,
	}
}

func (s *Server) Run() {
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world")
	})

	// TODO input addr
	if err := s.app.Listen(":8080"); err != nil {
		s.logger.Fatal("server crash", zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	s.app.Shutdown()
}

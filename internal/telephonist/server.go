package telephonist

import (
	"github.com/famusovsky/go-rufkian/internal/telephonist/walkietalkie"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type server struct {
	app          *fiber.App
	addr         string
	logger       *zap.Logger
	walkieTalkie walkietalkie.IController
}

// TODO instead of addr, input a normal config
func NewServer(logger *zap.Logger, db sqlx.Ext, addr string) IServer {
	return &server{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				logger.Error(
					"route handle",
					zap.String("url", c.OriginalURL()),
					zap.String("method", c.Method()),
					zap.Error(err),
				)
				// TODO more appropriate error
				return c.SendStatus(fiber.StatusNotFound)
			},
		}),
		addr:         addr,
		logger:       logger,
		walkieTalkie: walkietalkie.New(db, logger),
	}
}

func (s *server) Run() {
	s.initRouter()

	if err := s.app.Listen(s.addr); err != nil {
		s.logger.Fatal("server crash", zap.Error(err))
	}
}

func (s *server) Shutdown() {
	s.app.Shutdown()
}

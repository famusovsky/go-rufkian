package companion

import (
	"html/template"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/pkg/cookie"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type server struct {
	app           *fiber.App
	addr          string
	logger        *zap.Logger
	dbClient      database.IClient
	cookieHandler cookie.IHandler
}

// TODO instead of addr, input a normal config
func NewServer(logger *zap.Logger, db sqlx.Ext, addr string) IServer {
	engine := html.New("./ui/views", ".html")
	engine.AddFunc(
		"unescape", func(s string) template.HTML {
			return template.HTML(s)
		},
	)

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
			Views: engine,
		}),
		dbClient:      database.NewClient(db, logger),
		addr:          addr,
		logger:        logger,
		cookieHandler: cookie.NewHandler(cookieName),
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

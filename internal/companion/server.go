package companion

import (
	"html/template"

	"github.com/famusovsky/go-rufkian/internal/companion/auth"
	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/dialog"
	"github.com/famusovsky/go-rufkian/internal/companion/dictionary"
	"github.com/famusovsky/go-rufkian/internal/companion/proxy"
	"github.com/famusovsky/go-rufkian/internal/companion/user"
	"github.com/famusovsky/go-rufkian/pkg/cookie"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type server struct {
	app  *fiber.App
	addr string

	logger        *zap.Logger
	dbClient      database.IClient
	cookieHandler cookie.IHandler

	dialogHandlers     dialogHandlers
	authHandlers       authHandlers
	proxyHandlers      proxyHandlers
	dictionaryHandlers dictionaryHandlers
	userHandlers       userHandlers
}

// TODO instead of addr, input a normal config
func NewServer(logger *zap.Logger, db sqlx.Ext, addr string) (IServer, error) {
	engine := html.New("./ui/views", ".html")
	engine.AddFunc(
		"unescape", func(s string) template.HTML {
			return template.HTML(s)
		},
	)

	dbClient := database.NewClient(db, logger)
	cookieHandler := cookie.NewHandler(cookieName)

	res := &server{
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
		addr:               addr,
		dbClient:           dbClient,
		cookieHandler:      cookieHandler,
		logger:             logger,
		authHandlers:       auth.NewHandlers(dbClient, cookieHandler, logger),
		dialogHandlers:     dialog.NewHandlers(dbClient, logger),
		proxyHandlers:      proxy.NewHandlers(logger),
		dictionaryHandlers: dictionary.NewHandlers(dbClient, logger),
		userHandlers:       user.NewHandlers(dbClient, logger),
	}

	return res, nil
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

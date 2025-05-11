package telephonist

import (
	"time"

	"github.com/famusovsky/go-rufkian/internal/telephonist/translator"
	"github.com/famusovsky/go-rufkian/internal/telephonist/walkietalkie"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type server struct {
	app             *fiber.App
	addr            string
	logger          *zap.Logger
	walkieTalkie    walkietalkie.IController
	ticker          *time.Ticker
	companionClient *resty.Client
}

// TODO instead of addr, input a normal config
func NewServer(logger *zap.Logger, db sqlx.Ext, addr, companionURL, yaFolderID, yaTranslateKey string) IServer {
	return &server{
		app: fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				logger.Error(
					"route handle",
					zap.String("url", c.OriginalURL()),
					zap.String("method", c.Method()),
					zap.Error(err),
				)
				return c.SendStatus(fiber.StatusNotFound)
			},
		}),
		addr:         addr,
		logger:       logger,
		walkieTalkie: walkietalkie.New(db, logger, translator.NewYaClient(yaFolderID, yaTranslateKey)),
		ticker:       time.NewTicker(time.Minute),
		companionClient: resty.New().
			SetDisableWarn(true).
			SetAllowMethodGetPayload(true).
			SetContentLength(true).
			SetBaseURL(companionURL).
			SetTimeout(3 * time.Second),
	}
}

func (s *server) Run() {
	s.initRouter()

	go func() {
		for range s.ticker.C {
			s.walkieTalkie.CleanUp()
		}
	}()

	if err := s.app.Listen(s.addr); err != nil {
		s.logger.Fatal("server crash", zap.Error(err))
	}
}

func (s *server) Shutdown() {
	s.ticker.Stop()
	s.app.Shutdown()
}

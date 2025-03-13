package dictionary

import (
	"errors"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type handlers struct {
	dbClient database.IClient
	logger   *zap.Logger
}

func NewHandlers(
	dbClient database.IClient,
	logger *zap.Logger,
) IHandlers {
	res := &handlers{
		dbClient: dbClient,
		logger:   logger,
	}

	return res
}

func (h *handlers) DictionaryPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting user's dictionary")

	user, _ := middleware.UserFromCtx(c)

	words, err := h.dbClient.GetUserWords(user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
	}

	return c.Render("dictionary", fiber.Map{
		"words": words,
	}, "layouts/base")
}

func (h *handlers) WordPage(c *fiber.Ctx) error {
	word := c.Params("word", "deutsch")
	previousPage := c.Query("previous_page")
	h.logger.Info("previous_page", zap.String("addr", previousPage))

	return c.Render("word", fiber.Map{
		"previousPage": previousPage,
		"word":         word,
	}, "layouts/base")
}

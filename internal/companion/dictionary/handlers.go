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

	user, _ := middleware.UserFromCtx(c)

	inDictionary, err := h.dbClient.CheckUserWord(user.ID, word)
	if err != nil {
		return render.ErrPage(c, fiber.StatusInternalServerError, err)
	}

	return c.Render("word", fiber.Map{
		"previousPage": previousPage,
		"word":         word,
		"inDictionary": inDictionary,
	}, "layouts/base")
}

func (h *handlers) AddWord(c *fiber.Ctx) error {
	word := c.Params("word")
	user, ok := middleware.UserFromCtx(c)
	if !ok || word == "" {
		return render.ErrToResult(c, fiber.ErrBadRequest)
	}

	if err := h.dbClient.AddWordToUser(user.ID, word); err != nil {
		return render.ErrToResult(c, fiber.ErrInternalServerError)
	}
	return c.Next()
}

func (h *handlers) DeleteWord(c *fiber.Ctx) error {
	word := c.Params("word")
	user, ok := middleware.UserFromCtx(c)
	if !ok || word == "" {
		return render.ErrToResult(c, fiber.ErrBadRequest)
	}

	if err := h.dbClient.DeleteWordFromUser(user.ID, word); err != nil {
		return render.ErrToResult(c, fiber.ErrInternalServerError)
	}
	return c.Next()
}

package dictionary

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/famusovsky/go-rufkian/pkg/apkg"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type handlers struct {
	dbClient database.IClient
	logger   *zap.Logger
	regular  *regexp.Regexp
}

func NewHandlers(
	dbClient database.IClient,
	logger *zap.Logger,
) IHandlers {
	regular, _ := regexp.Compile("[^a-zA-Z0-9]+")
	res := &handlers{
		dbClient: dbClient,
		logger:   logger,
		regular:  regular,
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

	if len(words) == 0 {
		return c.Render("dictionary", fiber.Map{
			"empty": true,
		}, "layouts/base")
	}

	return c.Render("dictionary", fiber.Map{
		"words":      words,
		"userHasKey": user.Key,
	}, "layouts/base")
}

func (h *handlers) WordPage(c *fiber.Ctx) error {
	word := c.Params("word", "deutsch")
	previousPage := c.Query("previous_page")

	word = h.regular.ReplaceAllString(word, "")
	word = strings.ToLower(word)
	h.logger.Info("word", zap.String("w", word))

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

	return c.SendStatus(fiber.StatusNoContent)
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

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *handlers) GetApkg(c *fiber.Ctx) error {
	errWrap := errors.New("error while exporting user's dictionary to apkg")

	user, _ := middleware.UserFromCtx(c)

	words, err := h.dbClient.GetUserWords(user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
	}

	notes := make([]apkg.SimpleNote, 0, len(words))
	for _, word := range words {
		notes = append(notes, apkg.SimpleNote{
			Front: word,
			Back:  "default back", // TODO
		})
	}

	res, err := apkg.Convert(apkg.NewSimpleAnki(notes))
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
	}

	c.Response().Header.Add(fiber.HeaderContentDisposition, `attachment; filename="rufkian.apkg"`)
	return c.SendStream(bytes.NewReader(res))
}

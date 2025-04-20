package dictionary

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/famusovsky/go-rufkian/internal/model"
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
	slices.Sort(words)

	hash := md5.Sum([]byte(strings.Join(words, "")))
	stringHash := hex.EncodeToString(hash[:])
	needUpdate, err := h.dbClient.CheckDictionaryNeedUpdate(user.ID, stringHash)
	if err != nil {
		h.logger.Error("chech if dictionary needs update", zap.Error(err), zap.String("user_id", user.ID))
		needUpdate = true
	}

	var res []byte
	if !needUpdate {
		dict, err := h.dbClient.GetDictionary(user.ID)
		if err != nil {
			return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
		}
		res = dict.Apkg
	} else {
		notes := make([]apkg.SimpleNote, 0, len(words))
		for _, word := range words {
			// TODO do a tx or smth
			info, err := h.dbClient.GetWord(word)
			if err != nil {
				continue
			}

			notes = append(notes, apkg.SimpleNote{
				Front: word,
				Back:  info,
			})
		}

		converted, err := apkg.Convert(apkg.NewSimpleAnki(notes))
		if err != nil {
			return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
		}
		res = converted

		go func() {
			dict := model.Dictionary{
				UserID: user.ID,
				Hash:   stringHash,
				Apkg:   res,
			}
			if err := h.dbClient.UpdateDictionary(dict); err != nil {
				h.logger.Error("update user's dictionary", zap.Error(err), zap.String("user_id", user.ID))
			}
		}()
	}

	c.Response().Header.Add(fiber.HeaderContentDisposition, `attachment; filename="rufkian.apkg"`)
	return c.SendStream(bytes.NewReader(res))
}

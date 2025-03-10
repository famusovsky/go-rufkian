package dialog

import (
	"bytes"
	"errors"
	"html/template"
	"time"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type handlers struct {
	tmpl     *template.Template
	dbClient database.IClient
	logger   *zap.Logger
}

func NewHandlers(
	dbClient database.IClient,
	logger *zap.Logger,
) (IHandlers, error) {
	res := handlers{
		tmpl:     template.New(""),
		dbClient: dbClient,
		logger:   logger,
	}

	_, historyErr := res.tmpl.New(tmplHistoryName).Parse(tmplHistoryText)
	_, dialogErr := res.tmpl.New(tmplDialogName).Parse(tmplDialogText)

	if err := errors.Join(historyErr, dialogErr); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *handlers) RenderPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting dialog")

	user, _ := middleware.UserFromCtx(c)
	dialogID := c.Params("id")

	dialog, err := h.dbClient.GetDialog(dialogID, user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, err)
	}

	var body bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&body, tmplDialogName, dialog.Messages); err != nil {
		return render.ErrPage(c, fiber.StatusInternalServerError, errors.Join(errWrap, err))
	}

	return c.Render("dialog", fiber.Map{
		"tbody":     body.String(),
		"startTime": dialog.StartTime,
	}, "layouts/base")
}

func (h *handlers) RenderHistoryPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting user's dialog history")

	user, _ := middleware.UserFromCtx(c)

	dialogs, err := h.dbClient.GetUserDialogs(user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, err)
	}

	type dialogView struct {
		StartTime time.Time
		FirstLine string
		ID        string
	}

	res := make([]dialogView, 0, len(dialogs))
	for _, dialog := range dialogs {
		if dialog.Messages[0].Empty() {
			continue
		}
		res = append(res, dialogView{
			StartTime: dialog.StartTime,
			FirstLine: dialog.Messages[0].Content,
			ID:        dialog.ID,
		})
	}

	var body bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&body, tmplHistoryName, res); err != nil {
		return render.ErrPage(c, fiber.StatusInternalServerError, errors.Join(errWrap, err))
	}

	return c.Render("history", fiber.Map{
		"tbody": body.String(),
	}, "layouts/base")
}

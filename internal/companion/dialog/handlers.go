package dialog

import (
	"errors"
	"strings"
	"time"

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
	res := handlers{
		dbClient: dbClient,
		logger:   logger,
	}

	return &res
}

func (h *handlers) DialogPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting dialog")

	user, _ := middleware.UserFromCtx(c)
	dialogID := c.Params("id")

	dialog, err := h.dbClient.GetDialog(dialogID, user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
	}

	type lineView struct {
		Role        string
		Words       []string
		Translation *string
	}

	dialogView := struct {
		ID         string
		Translated bool
		Lines      []lineView
	}{
		ID:         dialogID,
		Translated: true,
	}

	dialogView.Lines = make([]lineView, len(dialog.Messages))
	for i, msg := range dialog.Messages {
		dialogView.Lines[i].Role = string(msg.Role)
		dialogView.Lines[i].Words = strings.Fields(msg.Content)
		dialogView.Lines[i].Translation = msg.Translation
		if msg.Translation == nil {
			dialogView.Translated = false
		}
	}

	return c.Render("dialog", fiber.Map{
		"dialog":    dialogView,
		"startTime": dialog.StartTime,
	}, "layouts/base")
}

func (h *handlers) HistoryPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting user's dialog history")

	user, _ := middleware.UserFromCtx(c)

	dialogs, err := h.dbClient.GetUserDialogs(user.ID)
	if err != nil {
		return render.ErrPage(c, fiber.StatusNotFound, errors.Join(errWrap, err))
	}

	type dialogView struct {
		StartTime time.Time
		FirstLine string
		ID        string
	}

	dialogViews := make([]dialogView, 0, len(dialogs))
	for _, dialog := range dialogs {
		if dialog.Messages[0].Empty() {
			continue
		}

		firstLine := dialog.Messages[0].Content
		if idx := strings.Index(firstLine, "."); idx > 0 {
			firstLine = firstLine[:idx]
		} else if idx := strings.Index(firstLine, "?"); idx > 0 {
			firstLine = firstLine[:idx]
		} else if idx := strings.Index(firstLine, "!"); idx > 0 {
			firstLine = firstLine[:idx]
		}

		dialogViews = append(dialogViews, dialogView{
			StartTime: dialog.StartTime,
			FirstLine: firstLine,
			ID:        dialog.ID,
		})
	}

	return c.Render("history", fiber.Map{
		"dialogs": dialogViews,
	}, "layouts/base")
}

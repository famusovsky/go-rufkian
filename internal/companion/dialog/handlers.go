package dialog

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/middleware"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/famusovsky/go-rufkian/internal/model"
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
	regular, _ := regexp.Compile("[^!,?,.]+.")

	res := handlers{
		dbClient: dbClient,
		logger:   logger,
		regular:  regular,
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
		"startTime": dialog.StartTime.Format("02.01.2006 15:04"),
		"duration":  fmt.Sprintf("минут: %d, секунд: %d", dialog.DurationS/60, dialog.DurationS%60),
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
		StartTime string
		FirstLine string
		ID        string
	}

	slices.SortFunc(dialogs, func(i, j model.Dialog) int {
		return j.StartTime.Compare(i.StartTime)
	})

	day := 24 * time.Hour
	positiveStreak := 0
	streakCancelled := false
	currentTime := time.Now().UTC().Truncate(day)
	dialogViews := make([]dialogView, 0, len(dialogs))

	for _, dialog := range dialogs {
		if dialog.Messages[0].Empty() {
			continue
		}

		if !streakCancelled {
			dialogTime := dialog.StartTime.UTC().Truncate(day)
			timeDiff := currentTime.Sub(dialogTime)
			if timeDiff <= day && (!user.HasTimeGoal() || dialog.DurationS/60 >= *user.TimeGoalM) {
				currentTime = dialogTime
				positiveStreak++
			} else {
				streakCancelled = true
			}
		}

		firstLine := h.regular.FindString(dialog.Messages[0].Content)

		dialogViews = append(dialogViews, dialogView{
			StartTime: dialog.StartTime.Format("02.01.2006 15:04"),
			FirstLine: firstLine,
			ID:        dialog.ID,
		})
	}

	return c.Render("history", fiber.Map{
		"dialogs":        dialogViews,
		"userHasKey":     user.HasKey(),
		"showCallButton": string(c.Context().UserAgent()) == "rufkian",
		"daysWithGoal":   positiveStreak,
	}, "layouts/base")
}

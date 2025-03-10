package companion

import (
	"bytes"
	"errors"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AUTH

func (s *server) RenderAuthPage(c *fiber.Ctx) error {
	return c.Render("auth", fiber.Map{}, "layouts/mini")
}

// BASE

func (s *server) RenderHistoryPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting user's dialog history")

	user, _ := UserFromCtx(c)

	dialogs, err := s.dbClient.GetUserDialogs(user.ID)
	if err != nil {
		return s.renderErr(c, fiber.StatusNotFound, err)
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

	q := `{{range .}}<tr hx-get={{printf "/dialog/%d" .ID }} hx-target="body">
	<td>{{.StartTime}}</td>
	<td>{{.FirstLine}}</td>
	</tr>{{end}}`
	t := template.Must(template.New("").Parse(q))
	var body bytes.Buffer
	if err := t.Execute(&body, res); err != nil {
		return s.renderErr(c, fiber.StatusInternalServerError, errors.Join(errWrap, err))
	}

	return c.Render("history", fiber.Map{
		"tbody": body.String(),
	}, "layouts/base")
}

func (s *server) RenderDialogPage(c *fiber.Ctx) error {
	errWrap := errors.New("error while getting dialog")

	user, _ := UserFromCtx(c)
	dialogID := c.Get("id")

	dialog, err := s.dbClient.GetDialog(dialogID, user.ID)
	if err != nil {
		return s.renderErr(c, fiber.StatusNotFound, err)
	}

	// XXX probably should not use New every time
	q := `{{range .}}<tr>
	<td>{{.Role}}</td>
	<td>{{.Content}}</td>
	</tr>{{end}}`
	t := template.Must(template.New("").Parse(q))
	var body bytes.Buffer
	if err := t.Execute(&body, dialog.Messages); err != nil {
		return s.renderErr(c, fiber.StatusInternalServerError, errors.Join(errWrap, err))
	}

	return c.Render("dialog", fiber.Map{
		"tbody":     body.String(),
		"startTime": dialog.StartTime,
	}, "layouts/base")
}

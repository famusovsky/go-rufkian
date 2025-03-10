package render

import (
	"github.com/gofiber/fiber/v2"
)

// XXX maibe should rename package

// TODO change from ... to Option

func ErrToResult(c *fiber.Ctx, err error, name ...string) error {
	result := "#result"
	if len(name) != 0 {
		result = name[0]
	}
	c.Set("HX-Retarget", result)
	return c.SendString(err.Error())
}

func ErrPage(c *fiber.Ctx, status int, err error, layout ...string) error {
	l := "layouts/base"
	if len(layout) != 0 {
		l = layout[0]
	}
	return c.Render("error", fiber.Map{
		"status":  status,
		"errText": err.Error(),
	}, l)
}

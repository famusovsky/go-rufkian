package companion

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TODO change from ... to Option

func (s *server) setErrToResult(c *fiber.Ctx, err error, name ...string) error {
	result := "#result"
	if len(name) != 0 {
		result = name[0]
	}
	s.logger.Info("rendered error in #result", zap.Error(err))
	c.Set("HX-Retarget", result)
	return c.SendString(err.Error())
}

func (s *server) renderErr(c *fiber.Ctx, status int, err error, layout ...string) error {
	l := "layouts/base"
	if len(layout) != 0 {
		l = layout[0]
	}
	s.logger.Info("rendered error page", zap.Error(err))
	return c.Render("error", fiber.Map{
		"status":  status,
		"errText": err.Error(),
	}, l)
}

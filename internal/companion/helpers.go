package companion

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (s *server) setErrToResult(c *fiber.Ctx, err error, name ...string) error {
	result := "#result"
	if len(name) != 0 {
		result = name[0]
	}
	s.logger.Info("rendered error in #result", zap.Error(err))
	c.Set("HX-Retarget", result)
	return c.SendString(err.Error())
}

package companion

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (s *server) hello(c *fiber.Ctx) error {
	user, ok := UserFromCtx(c)
	if !ok {
		return c.SendString("NOT SIGNED IN")
	}
	return c.SendString(fmt.Sprintf("HELLO, %s !!!", user.Email))
}

func (s *server) favicon(c *fiber.Ctx) error {
	return c.SendFile("ui/static/favicon.ico")
}

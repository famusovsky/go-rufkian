package telephonist

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

// FIXME simplify this file, a lot of boilerplate here

type postRequestPayload struct {
	Key   string `json:"key"`
	Input string `json:"input"`
}

type postResponsePayload struct {
	Answer string `json:"answer,omitempty"`
	Status string `json:"status,omitempty"`
}

func (s *Server) Post(c *fiber.Ctx) error {
	var payload postRequestPayload
	if err := c.BodyParser(&payload); err != nil {
		s.logger.Warn("bad post request", zap.Error(err))
		return c.JSON(postResponsePayload{
			Status: utils.StatusMessage(fiber.StatusBadRequest),
		})
	}
	answer := s.walkieTalkie.Talk(payload.Key, payload.Input)

	return c.JSON(postResponsePayload{
		Answer: answer,
		Status: utils.StatusMessage(fiber.StatusOK),
	})
}

type deleteRequstPayload struct {
	Key string `json:"key"`
}

type deleteResponsePayload struct {
	ID     string `json:"dialog_id,omitempty"`
	Status string `json:"status,omitempty"`
}

func (s *Server) Delete(c *fiber.Ctx) error {
	var payload deleteRequstPayload
	if err := c.BodyParser(&payload); err != nil {
		s.logger.Warn("bad post request", zap.Error(err))
		return c.JSON(postResponsePayload{
			Status: utils.StatusMessage(fiber.StatusBadRequest),
		})
	}

	id, err := s.walkieTalkie.Stop(payload.Key)
	if err != nil {
		return c.JSON(deleteResponsePayload{
			Status: err.Error(),
		})
	}

	return c.JSON(deleteResponsePayload{
		ID:     id,
		Status: utils.StatusMessage(fiber.StatusOK),
	})
}

func (s *Server) Ping(c *fiber.Ctx) error {
	s.logger.Info("ping pong")
	return c.SendString("pong")
}

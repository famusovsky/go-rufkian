package telephonist

import (
	"errors"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"
)

type requestPayload struct {
	UserID *string `json:"user_id"`
	Key    *string `json:"key"`
	Input  *string `json:"input"`
}

type postResponsePayload struct {
	Answer string `json:"answer,omitempty"`
	Status string `json:"status,omitempty"`
}

func (s *server) Post(c *fiber.Ctx) error {
	var payload requestPayload
	err := c.BodyParser(&payload)
	if err != nil {
		s.logger.Error("parse post request payload", zap.Error(err), zap.ByteString("payload", c.Body()))
		err = model.ErrWrongBodyFormat
	} else {
		if payload.UserID == nil {
			err = model.ErrEmptyUserID
		}
		if payload.Key == nil {
			err = errors.Join(err, model.ErrEmptyKey)
		}
		if payload.Input == nil {
			err = errors.Join(err, model.ErrEmptyInput)
		}
	}

	if err != nil {
		s.logger.Info("post request payload", zap.Error(err))
		return c.
			Status(fiber.StatusBadRequest).
			JSON(postResponsePayload{
				Status: err.Error(),
			})
	}

	return c.JSON(postResponsePayload{
		Answer: s.walkieTalkie.Talk(*payload.UserID, *payload.Key, *payload.Input),
		Status: utils.StatusMessage(fiber.StatusOK),
	})
}

type deleteResponsePayload struct {
	ID     string `json:"dialog_id,omitempty"`
	Status string `json:"status,omitempty"`
}

func (s *server) Delete(c *fiber.Ctx) error {
	var payload requestPayload
	err := c.BodyParser(&payload)
	if err != nil {
		err = model.ErrWrongBodyFormat
	} else {
		if payload.Key == nil {
			err = errors.Join(err, model.ErrEmptyKey)
		}
		if payload.UserID == nil {
			err = model.ErrEmptyUserID
		}
	}

	if err != nil {
		s.logger.Warn("delete request payload", zap.Error(err))
		return c.
			Status(fiber.StatusBadRequest).
			JSON(postResponsePayload{
				Status: err.Error(),
			})
	}

	id, err := s.walkieTalkie.Stop(*payload.UserID, *payload.Key)
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

package telephonist

import (
	"encoding/json"
	"fmt"

	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func (s *server) getUser(c *fiber.Ctx) (model.User, error) {
	req, err := adaptor.ConvertRequest(c, false)
	if err != nil {
		return model.User{}, fmt.Errorf("convert fiber context to http request: %w", err)
	}

	response, err := s.companionClient.R().
		SetCookies(req.Cookies()).
		Get("user")

	if err != nil {
		return model.User{}, err
	}

	var user model.User
	if err := json.Unmarshal(response.Bytes(), &user); err != nil {
		return model.User{}, fmt.Errorf("parse GET /user body: %w", err)
	}

	if user.Key == nil {
		return model.User{}, fmt.Errorf("user %s has no key", user.ID)
	}

	return user, nil
}

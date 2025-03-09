package companion

import (
	"errors"
	"fmt"
	"time"

	"github.com/badoux/checkmail"
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var (
	errWrap = errors.New("sign up user handler: %w")
	day     = 24 * time.Hour
	week    = 7 * day
)

// TODO unite signup and signing

func (s *server) SignUp(c *fiber.Ctx) error {
	c.Accepts("json")
	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return s.setErrToResult(c, errors.Join(errWrap, fmt.Errorf("parse request body to user: %w", err)))
	}

	if err := checkmail.ValidateFormat(user.Email); err != nil {
		return s.setErrToResult(c, errors.Join(errWrap, fmt.Errorf("validate email: %w", err)))
	}

	user, err := s.dbClient.AddUser(user)
	if err != nil {
		return s.setErrToResult(c, errors.Join(errWrap, fmt.Errorf("add user in db: %w", err)))
	}

	s.cookieHandler.Set(&c.Response().Header, time.Now().Add(week), userIDKey, user.ID)

	s.logger.Info("user signed up", zap.String("user", user.ID))

	c.Set("HX-Location", "/")
	return c.SendString("OK")
}

func (s *server) SignIn(c *fiber.Ctx) error {
	errWrap := errors.New("error while signing in user in api")
	var user model.User
	if err := c.BodyParser(&user); err != nil || user.Email == "" || user.Password == "" {
		return s.setErrToResult(c, errors.Join(errWrap, errors.Join(errWrap, fmt.Errorf(`request's body is wrong`))))
	}

	user, err := s.dbClient.GetUserByCredentials(user)
	if err != nil {
		return s.setErrToResult(c, errors.Join(errWrap, fmt.Errorf("get user from db by email: %w", err)))
	}

	s.cookieHandler.Set(&c.Response().Header, time.Now().Add(week), userIDKey, user.ID)

	s.logger.Info("user signed in", zap.String("user_id", user.ID))

	c.Set("HX-Location", "/")
	return c.SendString("OK")
}

func (s *server) signOut(c *fiber.Ctx) error {
	s.cookieHandler.Clear(&c.Response().Header)

	c.Set("HX-Refresh", "true")
	return c.SendString("")
}

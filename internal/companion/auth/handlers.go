package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/badoux/checkmail"
	"github.com/famusovsky/go-rufkian/internal/companion/database"
	"github.com/famusovsky/go-rufkian/internal/companion/render"
	"github.com/famusovsky/go-rufkian/internal/model"
	"github.com/famusovsky/go-rufkian/pkg/cookie"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	day  = 24 * time.Hour
	week = 7 * day
)

type handlers struct {
	dbClient      database.IClient
	cookieHandler cookie.IHandler
	logger        *zap.Logger
}

func NewHandlers(
	dbClient database.IClient,
	cookieHandler cookie.IHandler,
	logger *zap.Logger,
) IHandlers {
	return &handlers{
		dbClient:      dbClient,
		cookieHandler: cookieHandler,
		logger:        logger,
	}
}

// TODO unite signup and signing

func (h *handlers) SignUp(c *fiber.Ctx) error {
	errWrap := errors.New("sign up user: %w")
	c.Accepts("json")
	var user model.User

	if err := c.BodyParser(&user); err != nil {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("parse request body to user: %w", err)))
	}

	if err := checkmail.ValidateFormat(user.Email); err != nil {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("validate email: %w", err)))
	}

	if len(user.Password) < 8 {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("пароль должен быть длиннее 8 строк")))
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
		if err != nil {
			h.logger.Error("generate hashed password", zap.Error(err))
			return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("неизвестная ошибка")))
		}
		user.Password = string(hashedPassword)
	}

	user, err := h.dbClient.AddUser(user)
	if err != nil {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("add user in db: %w", err)))
	}

	h.cookieHandler.Set(&c.Response().Header, time.Now().Add(week), model.UserKey, user.ID)

	h.logger.Info("user signed up", zap.String("user", user.ID))

	c.Set("HX-Location", "/")
	return c.SendString("OK")
}

func (h *handlers) SignIn(c *fiber.Ctx) error {
	errWrap := errors.New("sign in user")
	var user model.User
	if err := c.BodyParser(&user); err != nil || user.Email == "" || user.Password == "" {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("request's body is wrong")))
	}

	user, err := h.dbClient.GetUserByCredentials(user)
	if err != nil {
		return render.ErrToResult(c, fmt.Errorf(errWrap.Error(), fmt.Errorf("get user from db by email: %w", err)))
	}

	h.cookieHandler.Set(&c.Response().Header, time.Now().Add(week), model.UserKey, user.ID)

	h.logger.Info("user signed in", zap.String("user_id", user.ID))

	c.Set("HX-Location", "/")
	return c.SendString("OK")
}

func (h *handlers) SignOut(c *fiber.Ctx) error {
	h.cookieHandler.Clear(&c.Response().Header)

	c.Set("HX-Refresh", "true")
	return c.SendString("")
}

func (h *handlers) AuthPage(c *fiber.Ctx) error {
	return c.Render("auth", fiber.Map{}, "layouts/mini")
}

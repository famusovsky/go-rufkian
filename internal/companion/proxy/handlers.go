package proxy

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"resty.dev/v3"
)

type handlers struct {
	restyClient *resty.Client
	logger      *zap.Logger
}

func NewHandlers(logger *zap.Logger) IHandlers {
	return &handlers{
		restyClient: resty.New(),
		logger:      logger,
	}
}

func (h *handlers) Woerter(c *fiber.Ctx) error {
	q := c.Params("q")
	if q == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	r, err := h.restyClient.R().Get(woerterURL + q)
	if err != nil {
		h.logger.Warn("GET woerter", zap.Error(err))
		return c.SendStatus(fiber.StatusNotFound)
	}

	if r.IsSuccess() {
		page, err := html.Parse(r.Body)
		defer r.Body.Close()

		if err != nil {
			h.logger.Error("parse body from woerter", zap.Error(err))
			return c.SendStatus(fiber.StatusNotFound)
		}

		var found bool
		var f func(*html.Node)
		f = func(n *html.Node) {
			attrs := n.Attr
			for _, attr := range attrs {
				if attr.Key == "class" && attr.Val == woerterWordElementID {
					// TODO remove images from the html, add corresponding css
					err = html.Render(c, n)
					found = true
					return
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(page)

		if err != nil {
			h.logger.Error("render woerter word html", zap.Error(err))
			return c.SendStatus(fiber.StatusNotFound)
		}

		if !found {
			return c.SendStatus(fiber.StatusNotFound)
		}

		return c.SendStatus(fiber.StatusOK)
	}

	return c.SendStatus(r.StatusCode())
}

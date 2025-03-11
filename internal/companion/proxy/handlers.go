package proxy

import (
	"slices"
	"strings"
	"sync"

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

		definition := findArticle(page)

		if definition == nil {
			return c.SendStatus(fiber.StatusNotFound)
		}

		removeExcessiveHighChildren(definition)
		cleanInsides(definition)

		err = html.Render(c, definition)

		if err != nil {
			h.logger.Error("write word definition and style from buffer to ctx", zap.Error(err))
			return fiber.ErrInternalServerError
		}

		return c.SendStatus(fiber.StatusOK)
	}

	return c.SendStatus(r.StatusCode())
}

func findArticle(n *html.Node) *html.Node {
	var res *html.Node
	var f func(*html.Node)
	wg := &sync.WaitGroup{}
	f = func(n *html.Node) {
		defer wg.Done()
		if n.Data == "article" {
			res = n.FirstChild
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			wg.Add(1)
			go f(c)
		}
	}
	wg.Add(1)
	go f(n)
	wg.Wait()

	return res
}

func removeExcessiveHighChildren(n *html.Node) {
	excessiveHighNodes := []*html.Node{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		var ok bool
		for _, attr := range c.Attr {
			if attr.Key == "class" && slices.Contains(classesToReturn, attr.Val) {
				ok = true
				break
			}
		}
		if !ok {
			excessiveHighNodes = append(excessiveHighNodes, c)
		}
	}
	for _, excessiveNode := range excessiveHighNodes {
		n.RemoveChild(excessiveNode)
	}
}

func cleanInsides(n *html.Node) {
	wg := sync.WaitGroup{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		defer wg.Done()

		if n.FirstChild == nil {
			return
		}

		toExclude := []*html.Node{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if checkIfMustExcludeNode(c) {
				toExclude = append(toExclude, c)
			} else {
				wg.Add(1)
				go f(c)
			}

			c.Attr = slices.DeleteFunc(c.Attr, func(attr html.Attribute) bool {
				return !slices.Contains(attributesToSave, attr.Key)
			})
		}

		for _, c := range toExclude {
			n.RemoveChild(c)
		}
		if checkIfMustPopNode(n) {
			popNode(n)
		}
	}
	wg.Add(1)
	go f(n)
	wg.Wait()
}

func checkIfMustExcludeNode(n *html.Node) bool {
	if n.FirstChild == nil && n.Type == html.TextNode && (n.Data == "\n" || n.Data == "") {
		return true
	}
	if slices.Contains(elementsWithTypeToExclude, n.Data) {
		return true
	}
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			for _, v := range strings.Split(attr.Val, " ") {
				if slices.Contains(elementsWithClassToExclude, v) {
					return true
				}
			}
		}
		if attr.Key == "lang" && !slices.Contains(languagesToSave, attr.Val) {
			return true
		}
	}
	return false
}

func checkIfMustPopNode(n *html.Node) bool {
	return slices.Contains(elementsWithTypeToPop, n.Data)
}

func popNode(n *html.Node) {
	if n.Parent == nil {
		return
	}
	children := []*html.Node{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	for _, c := range children {
		n.RemoveChild(c)
		n.Parent.InsertBefore(c, n)
	}
	n.Parent.RemoveChild(n)
}

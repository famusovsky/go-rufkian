package cookie

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type cookieHandler struct {
	instance   *securecookie.SecureCookie
	cookieName string
}

func NewHandler(cookieName string) IHandler {
	hashKey, blockKey := make([]byte, 32), make([]byte, 16)
	rand.Read(hashKey)
	rand.Read(blockKey)

	var s = securecookie.New(hashKey, blockKey)

	return &cookieHandler{
		instance:   s,
		cookieName: cookieName,
	}
}

func (c *cookieHandler) Set(a adder, expireDate time.Time, keyValue ...string) error {
	if len(keyValue) < 2 {
		return nil
	}

	cookieValue := make(map[string]string, len(keyValue)/2)

	for i := range len(keyValue) - 1 {
		key, value := keyValue[i], keyValue[i+1]
		cookieValue[key] = value
	}

	encodedCookieValue, err := c.instance.Encode(c.cookieName, cookieValue)

	if err != nil {
		return fmt.Errorf("encode cookie value: %v", err)
	}

	cookie := &http.Cookie{
		Name:    c.cookieName,
		Value:   encodedCookieValue,
		Path:    "/",
		Secure:  true,
		Expires: expireDate,
	}

	setCookie(a, cookie)

	return nil
}

func (c *cookieHandler) Read(r requestWithCookies) (map[string]string, error) {
	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		return nil, fmt.Errorf("read cookie from http request: %v", err)
	}

	res := make(map[string]string)
	if cookie.Value != "" {
		if err := c.instance.Decode(c.cookieName, cookie.Value, &res); err != nil {
			return nil, fmt.Errorf("decode cookie value: %v", err)
		}
	}

	return res, nil
}

func (c *cookieHandler) Clear(a adder) {
	setCookie(a, &http.Cookie{
		Name: c.cookieName,
		Path: "/",
	})
}

func setCookie(a adder, cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		a.Add("Set-Cookie", v)
	}
}

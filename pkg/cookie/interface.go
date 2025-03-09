package cookie

import (
	"net/http"
	"time"
)

type IHandler interface {
	Set(a adder, expireDate time.Time, keyValue ...string) error
	Clear(a adder)
	Read(r requestWithCookies) (map[string]string, error)
}

type adder interface {
	Add(key, value string)
}

type requestWithCookies interface {
	Cookie(name string) (*http.Cookie, error)
}

package cookie

import (
	"net/http"
	"time"
)

//go:generate mockgen -package cookie -mock_names IHandler=HandlerMock,adder=adderMock,requestWithCookies=requestMock -source ./interface.go -typed -destination interface.mock.gen.go
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

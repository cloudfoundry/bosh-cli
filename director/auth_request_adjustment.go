package director

import (
	"net/http"
	"net/url"
)

type RedirectFunc func(*http.Request, []*http.Request) error

type AuthRequestAdjustment struct {
	authFunc func(bool) (string, error)
	username string
	password string
}

func NewAuthRequestAdjustment(
	authFunc func(bool) (string, error),
	username string,
	password string,
) AuthRequestAdjustment {
	return AuthRequestAdjustment{
		authFunc: authFunc,
		username: username,
		password: password,
	}
}

func (a AuthRequestAdjustment) NeedsReadjustment(resp *http.Response) bool {
	return resp.StatusCode == 401
}

func (a AuthRequestAdjustment) Adjust(req *http.Request, retried bool) error {
	if len(a.username) > 0 {
		req.URL.User = url.UserPassword(a.username, a.password)
	} else if a.authFunc != nil {
		authHeader, err := a.authFunc(retried)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", authHeader)
	}

	return nil
}

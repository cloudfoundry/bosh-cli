package uaa

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AccessTokenSession struct {
	uaa         UAA
	token       AccessToken
	config      ConfigUpdater
	environment string
}

func NewAccessTokenSession(uaa UAA, token AccessToken, config ConfigUpdater, environment string) *AccessTokenSession {
	return &AccessTokenSession{
		uaa:         uaa,
		token:       token,
		config:      config,
		environment: environment,
	}
}

// TokenFunc retrieves new access token on first time use
// instead of using existing access token optimizing for token
// being valid for a longer period of time. Subsequent calls
// will reuse access token until it's time for it to be refreshed.
func (s *AccessTokenSession) TokenFunc(retried bool) (string, error) {
	if !s.token.IsValid() || retried {
		refreshToken, refreshable := s.token.(RefreshableAccessToken)
		if !refreshable {
			return "", errors.New("not a refresh token")
		}

		tokenResp, err := s.uaa.RefreshTokenGrant(refreshToken.RefreshValue())
		if err != nil {
			return "", bosherr.WrapError(err, "refreshing token")
		}

		s.token = tokenResp
		if err = s.saveTokenCreds(); err != nil {
			return "", err
		}
	}

	return s.token.Type() + " " + s.token.Value(), nil
}

type ConfigUpdater interface {
	UpdateConfigWithToken(environment string, token AccessToken) error
	Save() error
}

func (s *AccessTokenSession) saveTokenCreds() error {
	return s.config.UpdateConfigWithToken(s.environment, s.token)
}

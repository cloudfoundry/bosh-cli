package uaa

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AccessTokenImpl struct {
	type_       string
	accessValue string
}

var _ AccessToken = &AccessTokenImpl{}

func (t AccessTokenImpl) Type() string  { return t.type_ }
func (t AccessTokenImpl) Value() string { return t.accessValue }
func (t AccessTokenImpl) IsValid() bool { return t.type_ != "" && t.accessValue != "" }

type RefreshableAccessTokenImpl struct {
	accessToken  AccessToken
	refreshValue string
}

var _ RefreshableAccessToken = &RefreshableAccessTokenImpl{}

func (t RefreshableAccessTokenImpl) Type() string  { return t.accessToken.Type() }
func (t RefreshableAccessTokenImpl) Value() string { return t.accessToken.Value() }
func (t RefreshableAccessTokenImpl) IsValid() bool { return t.accessToken.IsValid() }

func (t RefreshableAccessTokenImpl) RefreshValue() string {
	return t.refreshValue
}

type TokenInfo struct {
	Username  string   `json:"user_name"` // e.g. "admin",
	Scopes    []string `json:"scope"`     // e.g. ["openid","bosh.admin"]
	ExpiredAt int      `json:"exp"`
	// ...snip...
}

func NewTokenInfoFromValue(value string) (TokenInfo, error) {
	var info TokenInfo

	segments := strings.Split(value, ".")
	if len(segments) != 3 {
		return info, bosherr.Error("Expected token value to have 3 segments")
	}

	bytes, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return info, bosherr.WrapError(err, "Decoding token info")
	}

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return info, bosherr.WrapError(err, "Unmarshaling token info")
	}

	return info, nil
}

func NewAccessToken(accessValueType, accessValue string) AccessToken {
	return AccessTokenImpl{
		type_:       accessValueType,
		accessValue: accessValue,
	}
}

func NewRefreshableAccessToken(accessValueType, accessValue, refreshValue string) RefreshableAccessToken {
	if len(refreshValue) == 0 {
		panic("Expected non-empty refresh token value")
	}

	return &RefreshableAccessTokenImpl{
		accessToken: AccessTokenImpl{
			type_:       accessValueType,
			accessValue: accessValue,
		},
		refreshValue: refreshValue,
	}
}

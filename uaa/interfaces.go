package uaa

//go:generate counterfeiter . UAA

type UAA interface {
	Prompts() ([]Prompt, error)

	RefreshTokenGrant(string) (AccessToken, error)
	ClientCredentialsGrant() (AccessToken, error)
	OwnerPasswordCredentialsGrant([]PromptAnswer) (AccessToken, error)
}

//go:generate counterfeiter . Token

// Token is a plain token with a value.
type Token interface {
	Type() string
	Value() string
	IsValid() bool
}

//go:generate counterfeiter . AccessToken

// AccessToken is purely an access token. It does not contain a refresh token and
// cannot be refreshed for another token.
type AccessToken interface {
	Token
}

//go:generate counterfeiter . RefreshableAccessToken

// RefreshableAccessToken is an access token with a refresh token that can be used
// to get another access token.
type RefreshableAccessToken interface {
	AccessToken
	RefreshValue() string
}

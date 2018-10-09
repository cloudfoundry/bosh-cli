package uaa

type ClientTokenSession struct {
	uaa   UAA
	token AccessToken
}

func NewClientTokenSession(uaa UAA) *ClientTokenSession {
	return &ClientTokenSession{uaa: uaa}
}

func (c *ClientTokenSession) TokenFunc(retried bool) (string, error) {
	if c.token == nil || retried {
		token, err := c.uaa.ClientCredentialsGrant()
		if err != nil {
			return "", err
		}

		c.token = token
	}

	return c.token.Type() + " " + c.token.Value(), nil
}

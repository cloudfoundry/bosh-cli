package director

import (
	"net/http"

	"github.com/cloudfoundry/bosh-utils/httpclient"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

//go:generate counterfeiter . Adjustment

type Adjustment interface {
	Adjust(req *http.Request, retried bool) error
	NeedsReadjustment(*http.Response) bool
}

//go:generate counterfeiter . AdjustedClient

type AdjustedClient interface {
	Do(*http.Request) (*http.Response, error)
}

type AdjustableClient struct {
	client     AdjustedClient
	adjustment Adjustment
}

func NewAdjustableClient(client AdjustedClient, adjustment Adjustment) AdjustableClient {
	return AdjustableClient{client: client, adjustment: adjustment}
}

func (c AdjustableClient) Do(req *http.Request) (*http.Response, error) {
	retried := req.Body != nil

	err := c.adjustment.Adjust(req, retried)
	if err != nil {
		return nil, err
	}

	originalBody, err := httpclient.MakeReplayable(req)
	if originalBody != nil {
		defer originalBody.Close()
	}
	if err != nil {
		return nil, bosherr.WrapError(err, "Making the request retryable")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if c.adjustment.NeedsReadjustment(resp) {
		resp.Body.Close()

		if req.GetBody != nil {
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, bosherr.WrapError(err, "Updating request body for retry")
			}
		}

		err := c.adjustment.Adjust(req, true)
		if err != nil {
			return nil, err
		}

		// Try one more time again after an adjustment
		return c.client.Do(req)
	}

	return resp, nil
}

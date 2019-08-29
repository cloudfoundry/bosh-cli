package cmd

import (
	"net/http"
	"net/http/httputil"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CurlCmd struct {
	ui            boshui.UI
	clientRequest boshdir.ClientRequest
}

func NewCurlCmd(ui boshui.UI, clientRequest boshdir.ClientRequest) CurlCmd {
	return CurlCmd{ui: ui, clientRequest: clientRequest}
}

func (c CurlCmd) Run(opts CurlOpts) error {
	updateReq := func(req *http.Request) {
		for _, h := range opts.Headers {
			req.Header.Add(h.Name, h.Value)
		}
	}

	var bodyBs []byte
	var resp *http.Response
	var respErr error

	switch opts.Method {
	case "GET":
		bodyBs, resp, respErr = c.clientRequest.RawGet(opts.Args.Path, nil, updateReq)

	case "POST":
		bodyBs, resp, respErr = c.clientRequest.RawPost(opts.Args.Path, opts.Body.Bytes, updateReq)

	case "PUT":
		bodyBs, resp, respErr = c.clientRequest.RawPut(opts.Args.Path, opts.Body.Bytes, updateReq)

	case "DELETE":
		if len(opts.Headers) > 0 {
			return bosherr.Errorf("Expected no headers")
		}
		bodyBs, resp, respErr = c.clientRequest.RawDelete(opts.Args.Path)

	default:
		return bosherr.Errorf("Unknown method '%s'", opts.Method)
	}

	if opts.ShowHeaders {
		restBs, serializeErr := httputil.DumpResponse(resp, false)
		if serializeErr != nil {
			return bosherr.WrapErrorf(serializeErr, "Dumping HTTP response")
		}

		c.ui.PrintBlock(restBs)
	}

	c.ui.PrintBlock(bodyBs)

	if respErr != nil {
		return bosherr.WrapErrorf(respErr, "Executing HTTP request")
	}

	return nil
}

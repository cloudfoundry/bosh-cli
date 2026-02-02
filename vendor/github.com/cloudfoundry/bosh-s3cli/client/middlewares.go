package client

import (
	"context"
	"fmt"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const acceptEncodingHeader = "Accept-Encoding"

type acceptEncodingKey struct{}

func getAcceptEncodingKey(ctx context.Context) (v string) {
	v, _ = middleware.GetStackValue(ctx, acceptEncodingKey{}).(string)
	return v
}

func setAcceptEncodingKey(ctx context.Context, value string) context.Context {
	return middleware.WithStackValue(ctx, acceptEncodingKey{}, value)
}

var dropAcceptEncodingHeader = middleware.FinalizeMiddlewareFunc("DropAcceptEncodingHeader",
	func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
		req, ok := in.Request.(*smithyhttp.Request)
		if !ok {
			return out, metadata, &v4.SigningError{Err: fmt.Errorf("unexpected request middleware type %T", in.Request)}
		}

		if ae := req.Header.Get(acceptEncodingHeader); len(ae) > 0 {
			ctx = setAcceptEncodingKey(ctx, ae)
			req.Header.Del(acceptEncodingHeader)
			in.Request = req
		}

		return next.HandleFinalize(ctx, in)
	},
)

var setAcceptEncodingHeader = middleware.FinalizeMiddlewareFunc("SetAcceptEncodingHeader",
	func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
		req, ok := in.Request.(*smithyhttp.Request)
		if !ok {
			return out, metadata, &v4.SigningError{Err: fmt.Errorf("unexpected request middleware type %T", in.Request)}
		}

		if ae := getAcceptEncodingKey(ctx); len(ae) > 0 {
			req.Header.Set(acceptEncodingHeader, ae)
			in.Request = req
		}

		return next.HandleFinalize(ctx, in)
	},
)

func AddFixAcceptEncodingMiddleware(stack *middleware.Stack) error {
	if _, ok := stack.Finalize.Get("Signing"); !ok {
		return nil
	}

	if err := stack.Finalize.Insert(dropAcceptEncodingHeader, "Signing", middleware.Before); err != nil {
		return err
	}

	if err := stack.Finalize.Insert(setAcceptEncodingHeader, "Signing", middleware.After); err != nil {
		return err
	}
	return nil
}

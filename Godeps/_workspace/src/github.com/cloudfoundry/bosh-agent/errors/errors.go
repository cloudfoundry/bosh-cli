package errors

import (
	"errors"
	"fmt"
)

type ShortenableError interface {
	error
	ShortError() string
}

type complexError struct {
	delegate error
	cause    error
}

func (e complexError) Error() string {
	return fmt.Sprintf("%s: %s", e.delegate.Error(), e.cause.Error())
}

func (e complexError) ShortError() string {
	var delegateMessage string
	if typedDelegate, ok := e.delegate.(ShortenableError); ok {
		delegateMessage = typedDelegate.ShortError()
	} else {
		delegateMessage = e.delegate.Error()
	}

	var causeMessage string
	if typedCause, ok := e.cause.(ShortenableError); ok {
		causeMessage = typedCause.ShortError()
	} else {
		causeMessage = e.cause.Error()
	}

	return fmt.Sprintf("%s: %s", delegateMessage, causeMessage)
}

func Error(msg string) error {
	return errors.New(msg)
}

func Errorf(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

func WrapError(cause error, msg string) error {
	return WrapComplexError(cause, Error(msg))
}

func WrapErrorf(cause error, msg string, args ...interface{}) error {
	return WrapComplexError(cause, Errorf(msg, args...))
}

func WrapComplexError(cause, delegate error) error {
	if cause == nil {
		cause = Error("<nil cause>")
	}

	return complexError{
		delegate: delegate,
		cause:    cause,
	}
}

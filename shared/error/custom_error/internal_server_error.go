package customErrors

import (
	"github.com/pkg/errors"
)

func NewInternalServerError(message string, code int) error {
	br := &internalServerError{
		CustomError: NewCustomError(nil, code, message),
	}
	stackErr := errors.WithStack(br)

	return stackErr
}

func NewInternalServerErrorWrap(err error, code int, message string) error {
	br := &internalServerError{
		CustomError: NewCustomError(err, code, message),
	}
	stackErr := errors.WithStack(br)

	return stackErr
}

type internalServerError struct {
	CustomError
}

type InternalServerError interface {
	CustomError
	IsInternalServerError() bool
}

func (i *internalServerError) IsInternalServerError() bool {
	return true
}

func IsInternalServerError(err error) bool {
	var internalErr InternalServerError
	//us, ok := grpc_errors.Cause(err).(InternalServerError)
	if errors.As(err, &internalErr) {
		return internalErr.IsInternalServerError()
	}

	return false
}
package errs

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppError struct {
	Code codes.Code
	Msg  string
	Err  error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Msg)
}

func InvalidArgument(msg string, err error) *AppError {
	return &AppError{Code: codes.InvalidArgument, Msg: msg, Err: err}
}

func Internal(msg string, err error) *AppError {
	return &AppError{Code: codes.Internal, Msg: msg, Err: err}
}

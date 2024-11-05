package cvvapi

import "fmt"

type ApiError struct {
	error
}

func NewApiError(e error) *ApiError {
	return &ApiError{e}
}

type NetworkError struct {
	error
}

func NewNetworkError(e error) *NetworkError {
	return &NetworkError{e}
}

type NotAuthenticatedError struct {
	error
}

func NewNotAuthenticatedError(e error) *NotAuthenticatedError {
	return &NotAuthenticatedError{e}
}

type HttpError struct {
	error
	StatusCode int
}

func NewHttpError(e error, statusCode int) *HttpError {
	return &HttpError{fmt.Errorf("%v: statusCode=%v", e.Error(), statusCode), statusCode}
}
package errors

import "net/http"

var (
	ErrBadRequest     = New(http.StatusBadRequest, "invalid request")
	ErrValidation     = New(http.StatusUnprocessableEntity, "validation error")
	ErrUnauthorized   = New(http.StatusUnauthorized, "unauthorized")
	ErrForbidden      = New(http.StatusForbidden, "forbidden")
	ErrNotFound       = New(http.StatusNotFound, "resource not found")
	ErrConflict       = New(http.StatusConflict, "resource already exists")
	ErrInternalServer = New(http.StatusInternalServerError, "internal server error")
)

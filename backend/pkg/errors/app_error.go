package errors

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Wrap(code int, message, internal string) *AppError {
	return &AppError{Code: code, Message: message}
}

// IsUnavailable reports whether err is a 503 AppError, meaning an
// infrastructure dependency (docker socket, pm2 daemon, ...) is missing.
func IsUnavailable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == 503
	}
	return false
}

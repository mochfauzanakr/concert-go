package exception

import "net/http"

type AppException struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppException) Error() string {
	return e.Message
}

func (e *AppException) Unwrap() error {
	return e.Err
}

func (e *AppException) Is(target error) bool {
	t, ok := target.(*AppException)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Message == t.Message
}

// Category Constructors
func BadRequest(message string) *AppException {
	return &AppException{
		Code:    http.StatusBadRequest,
		Message: message,
	}
}

func Unauthorized(message string) *AppException {
	return &AppException{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NotFound(message string) *AppException {
	return &AppException{
		Code:    http.StatusNotFound,
		Message: message,
	}
}

func Conflict(message string) *AppException {
	return &AppException{
		Code:    http.StatusConflict,
		Message: message,
	}
}

func Internal(message string, err error) *AppException {
	return &AppException{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

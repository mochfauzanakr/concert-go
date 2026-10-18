package exception_test

import (
	"errors"
	"net/http"
	"testing"

	"concert-go/internal/exception"
)

func TestAppException(t *testing.T) {
	err := exception.BadRequest("bad request message")
	if err.Error() != "bad request message" {
		t.Errorf("expected 'bad request message', got %s", err.Error())
	}
	if err.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", err.Code)
	}

	wrapped := exception.Internal("wrapper", errors.New("underlying"))
	if wrapped.Unwrap() == nil {
		t.Errorf("expected underlying error on unwrap")
	}

	target := exception.BadRequest("bad request message")
	if !errors.Is(err, target) {
		t.Errorf("expected errors.Is to match identical exceptions")
	}

	different := exception.Unauthorized("bad request message")
	if errors.Is(err, different) {
		t.Errorf("expected errors.Is not to match different codes")
	}
}

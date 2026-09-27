package usecase_test

import (
	"testing"

	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"
)

func TestHelloUsecase_GetGreeting(t *testing.T) {
	mockRepo := mock.NewHelloRepositoryMock()
	uc := usecase.NewHelloUsecase(mockRepo)

	msg, err := uc.GetGreeting()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if msg != "Hello from konserGo!" {
		t.Fatalf("expected 'Hello from konserGo!', got %s", msg)
	}
}

package usecase_test

import (
	"context"
	"testing"

	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"
)

func TestRoleUsecase_GetRoles(t *testing.T) {
	roleRepo := mock.NewRoleRepositoryMock()
	uc := usecase.NewRoleUsecase(roleRepo)

	roles, err := uc.GetRoles(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(roles) != 5 {
		t.Fatalf("expected 5 roles, got: %d", len(roles))
	}

	if roles[0].Name != "SUPER_ADMIN" {
		t.Errorf("expected first role to be SUPER_ADMIN, got: %s", roles[0].Name)
	}
}

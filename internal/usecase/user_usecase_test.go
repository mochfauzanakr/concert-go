package usecase_test

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

func setupUserUsecase() (*usecase.UserUsecase, *mockEmailSender) {
	userRepo := mock.NewUserRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, cfg)
	return uc, emailSender
}

func TestUserUsecase_InviteUser_Hierarchy(t *testing.T) {
	uc, _ := setupUserUsecase()
	ctx := context.Background()

	superAdminID := 1
	adminID := 2
	organizerID := 3
	buyerID := 4

	inviterID := uuid.New()

	t.Run("SUPER_ADMIN can invite ADMIN", func(t *testing.T) {
		res, err := uc.InviteUser(ctx, inviterID, &superAdminID, request.InviteUserRequest{
			Email: "newadmin@example.com",
			Role:  "ADMIN",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Role != "ADMIN" {
			t.Errorf("expected role ADMIN, got %s", res.Role)
		}
	})

	t.Run("SUPER_ADMIN can invite ORGANIZER", func(t *testing.T) {
		res, err := uc.InviteUser(ctx, inviterID, &superAdminID, request.InviteUserRequest{
			Email: "neworganizer@example.com",
			Role:  "ORGANIZER",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Role != "ORGANIZER" {
			t.Errorf("expected role ORGANIZER, got %s", res.Role)
		}
	})

	t.Run("SUPER_ADMIN cannot invite STAFF", func(t *testing.T) {
		_, err := uc.InviteUser(ctx, inviterID, &superAdminID, request.InviteUserRequest{
			Email: "staff@example.com",
			Role:  "STAFF",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("ADMIN can invite ORGANIZER", func(t *testing.T) {
		res, err := uc.InviteUser(ctx, inviterID, &adminID, request.InviteUserRequest{
			Email: "orgbyadmin@example.com",
			Role:  "ORGANIZER",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Role != "ORGANIZER" {
			t.Errorf("expected role ORGANIZER, got %s", res.Role)
		}
	})

	t.Run("ADMIN cannot invite ADMIN", func(t *testing.T) {
		_, err := uc.InviteUser(ctx, inviterID, &adminID, request.InviteUserRequest{
			Email: "anotheradmin@example.com",
			Role:  "ADMIN",
		})
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
	})

	t.Run("ORGANIZER can invite STAFF", func(t *testing.T) {
		res, err := uc.InviteUser(ctx, inviterID, &organizerID, request.InviteUserRequest{
			Email: "newstaff@example.com",
			Role:  "STAFF",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Role != "STAFF" {
			t.Errorf("expected role STAFF, got %s", res.Role)
		}
	})

	t.Run("USER cannot invite anyone", func(t *testing.T) {
		_, err := uc.InviteUser(ctx, inviterID, &buyerID, request.InviteUserRequest{
			Email: "somebody@example.com",
			Role:  "STAFF",
		})
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
	})

	t.Run("cannot invite duplicate pending invite", func(t *testing.T) {
		_, err := uc.InviteUser(ctx, inviterID, &superAdminID, request.InviteUserRequest{
			Email: "newadmin@example.com",
			Role:  "ADMIN",
		})
		if err == nil {
			t.Fatal("expected conflict error for pending invite, got nil")
		}
	})
}

func TestUserUsecase_AcceptInvite(t *testing.T) {
	ctx := context.Background()
	userRepo := mock.NewUserRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, cfg)

	rawToken := "valid-secret-token-12345"
	tokenHash := util.HashToken(rawToken)
	organizerRole, _ := roleRepo.FindByName(ctx, "ORGANIZER")

	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "invited@example.com",
		RoleID:    organizerRole.ID,
		InvitedBy: uuid.New(),
		TokenHash: tokenHash,
		Status:    entity.InviteStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	t.Run("mismatched passwords", func(t *testing.T) {
		_, err := uc.AcceptInvite(ctx, request.AcceptInviteRequest{
			Token:           rawToken,
			Name:            "Invited Person",
			Password:        "Password123!",
			ConfirmPassword: "Password999!",
		})
		if err == nil {
			t.Fatal("expected error on mismatched passwords")
		}
	})

	t.Run("weak password", func(t *testing.T) {
		_, err := uc.AcceptInvite(ctx, request.AcceptInviteRequest{
			Token:           rawToken,
			Name:            "Invited Person",
			Password:        "weak",
			ConfirmPassword: "weak",
		})
		if err == nil {
			t.Fatal("expected error on weak password")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := uc.AcceptInvite(ctx, request.AcceptInviteRequest{
			Token:           "wrong-token",
			Name:            "Invited Person",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		})
		if err == nil {
			t.Fatal("expected error on invalid token")
		}
	})

	t.Run("success activation", func(t *testing.T) {
		userRes, err := uc.AcceptInvite(ctx, request.AcceptInviteRequest{
			Token:           rawToken,
			Name:            "Invited Person",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		})
		if err != nil {
			t.Fatalf("expected successful activation, got %v", err)
		}
		if userRes.Email != "invited@example.com" {
			t.Errorf("expected email invited@example.com, got %s", userRes.Email)
		}
		if *userRes.RoleID != organizerRole.ID {
			t.Errorf("expected role ID %d, got %v", organizerRole.ID, userRes.RoleID)
		}

		// Verify user exists in repository and has status ACTIVE
		u, err := userRepo.FindByEmail(ctx, "invited@example.com")
		if err != nil || u == nil {
			t.Fatalf("expected user to be created in repository")
		}
		if u.Status != entity.UserStatusActive {
			t.Errorf("expected status ACTIVE, got %s", u.Status)
		}

		// Verify token cannot be reused
		_, err = uc.AcceptInvite(ctx, request.AcceptInviteRequest{
			Token:           rawToken,
			Name:            "Invited Person",
			Password:        "Password123!",
			ConfirmPassword: "Password123!",
		})
		if err == nil {
			t.Fatal("expected token to be consumed and rejected on second attempt")
		}
	})
}

func TestUserUsecase_ValidateInvite(t *testing.T) {
	ctx := context.Background()
	userRepo := mock.NewUserRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, cfg)

	rawToken := "valid-token-for-check"
	tokenHash := util.HashToken(rawToken)
	adminRole, _ := roleRepo.FindByName(ctx, "ADMIN")

	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "check@example.com",
		RoleID:    adminRole.ID,
		InvitedBy: uuid.New(),
		TokenHash: tokenHash,
		Status:    entity.InviteStatusPending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := uc.ValidateInvite(ctx, "invalid-token")
		if err == nil {
			t.Fatal("expected error on invalid token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := uc.ValidateInvite(ctx, "")
		if err == nil {
			t.Fatal("expected error on empty token")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		res, err := uc.ValidateInvite(ctx, rawToken)
		if err != nil {
			t.Fatalf("expected valid token result, got %v", err)
		}
		if res.Email != "check@example.com" {
			t.Errorf("expected email check@example.com, got %s", res.Email)
		}
		if res.Role != "ADMIN" {
			t.Errorf("expected role ADMIN, got %s", res.Role)
		}
	})
}

func TestUserUsecase_GetAllUsers(t *testing.T) {
	ctx := context.Background()
	userRepo := mock.NewUserRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, nil, cfg)

	adminRole, _ := roleRepo.FindByName(ctx, "ADMIN")
	buyerRole, _ := roleRepo.FindByName(ctx, "USER")

	_ = userRepo.Create(ctx, &entity.User{
		Name:     "Alice Admin",
		Email:    "alice@example.com",
		RoleID:   &adminRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})
	_ = userRepo.Create(ctx, &entity.User{
		Name:     "Bob Buyer",
		Email:    "bob@example.com",
		RoleID:   &buyerRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})
	_ = userRepo.Create(ctx, &entity.User{
		Name:     "Charlie Suspended",
		Email:    "charlie@example.com",
		RoleID:   &buyerRole.ID,
		Status:   entity.UserStatusSuspended,
		Provider: "email",
	})

	t.Run("all users pagination", func(t *testing.T) {
		pageRes, err := uc.GetAllUsers(ctx, request.GetUserListRequest{
			Page:     1,
			PageSize: 2,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pageRes.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(pageRes.Items))
		}
		if pageRes.Pagination.TotalRecords != 3 {
			t.Errorf("expected 3 total records, got %d", pageRes.Pagination.TotalRecords)
		}
		if pageRes.Pagination.TotalPage != 2 {
			t.Errorf("expected 2 total pages, got %d", pageRes.Pagination.TotalPage)
		}
		if pageRes.Pagination.NextPage == nil || *pageRes.Pagination.NextPage != 2 {
			t.Errorf("expected next page to be 2")
		}
	})

	t.Run("filter by search", func(t *testing.T) {
		pageRes, err := uc.GetAllUsers(ctx, request.GetUserListRequest{
			Search: "alice",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pageRes.Items) != 1 || pageRes.Items[0].Name != "Alice Admin" {
			t.Errorf("expected Alice Admin, got %+v", pageRes.Items)
		}
		if pageRes.Pagination.TotalRecords != 1 {
			t.Errorf("expected 1 record, got %d", pageRes.Pagination.TotalRecords)
		}
	})

	t.Run("filter by role", func(t *testing.T) {
		pageRes, err := uc.GetAllUsers(ctx, request.GetUserListRequest{
			RoleID: &adminRole.ID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pageRes.Items) != 1 || pageRes.Items[0].Email != "alice@example.com" {
			t.Errorf("expected alice@example.com, got %+v", pageRes.Items)
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		pageRes, err := uc.GetAllUsers(ctx, request.GetUserListRequest{
			Status: "SUSPENDED",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pageRes.Items) != 1 || pageRes.Items[0].Name != "Charlie Suspended" {
			t.Errorf("expected Charlie Suspended, got %+v", pageRes.Items)
		}
	})
}

func TestUserUsecase_UpdateUserStatus(t *testing.T) {
	ctx := context.Background()
	userRepo := mock.NewUserRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	emailSender := &mockEmailSender{}
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(userRepo, roleRepo, inviteRepo, emailSender, nil, cfg)

	superAdminRole, _ := roleRepo.FindByName(ctx, "SUPER_ADMIN")
	adminRole, _ := roleRepo.FindByName(ctx, "ADMIN")
	organizerRole, _ := roleRepo.FindByName(ctx, "ORGANIZER")

	superAdminID := uuid.New()
	admin1ID := uuid.New()
	admin2ID := uuid.New()
	organizerID := uuid.New()

	_ = userRepo.Create(ctx, &entity.User{
		ID:       superAdminID,
		Name:     "Super Admin",
		Email:    "superadmin@example.com",
		RoleID:   &superAdminRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})
	_ = userRepo.Create(ctx, &entity.User{
		ID:       admin1ID,
		Name:     "Admin One",
		Email:    "admin1@example.com",
		RoleID:   &adminRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})
	_ = userRepo.Create(ctx, &entity.User{
		ID:       admin2ID,
		Name:     "Admin Two",
		Email:    "admin2@example.com",
		RoleID:   &adminRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})
	_ = userRepo.Create(ctx, &entity.User{
		ID:       organizerID,
		Name:     "Event Organizer",
		Email:    "org@example.com",
		RoleID:   &organizerRole.ID,
		Status:   entity.UserStatusActive,
		Provider: "email",
	})

	t.Run("self-lockout rejected", func(t *testing.T) {
		_, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, admin1ID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusSuspended,
		})
		if err == nil {
			t.Fatal("expected error on self-lockout attempt")
		}
	})

	t.Run("invalid status string", func(t *testing.T) {
		_, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, organizerID, request.UpdateUserStatusRequest{
			Status: "BANNED",
		})
		if err == nil {
			t.Fatal("expected error on invalid status")
		}
	})

	t.Run("cannot suspend SUPER_ADMIN", func(t *testing.T) {
		_, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, superAdminID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusSuspended,
		})
		if err == nil {
			t.Fatal("expected error when trying to suspend super admin")
		}
	})

	t.Run("ADMIN cannot suspend another ADMIN", func(t *testing.T) {
		_, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, admin2ID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusSuspended,
		})
		if err == nil {
			t.Fatal("expected error when admin attempts to suspend another admin")
		}
	})

	t.Run("SUPER_ADMIN can suspend ADMIN", func(t *testing.T) {
		res, err := uc.UpdateUserStatus(ctx, superAdminID, &superAdminRole.ID, admin2ID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusSuspended,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != entity.UserStatusSuspended {
			t.Errorf("expected status SUSPENDED, got %s", res.Status)
		}
	})

	t.Run("ADMIN can suspend ORGANIZER", func(t *testing.T) {
		res, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, organizerID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusSuspended,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != entity.UserStatusSuspended {
			t.Errorf("expected status SUSPENDED, got %s", res.Status)
		}
	})

	t.Run("Reactivate user back to ACTIVE", func(t *testing.T) {
		res, err := uc.UpdateUserStatus(ctx, admin1ID, &adminRole.ID, organizerID, request.UpdateUserStatusRequest{
			Status: entity.UserStatusActive,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != entity.UserStatusActive {
			t.Errorf("expected status ACTIVE, got %s", res.Status)
		}
	})
}


package usecase_test

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"

	"github.com/google/uuid"
)

func TestOrganizerUsecase_UpsertProfile(t *testing.T) {
	ctx := context.Background()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, inviteRepo, roleRepo, sessionRepo, storageService)

	orgRole, _ := roleRepo.FindByName(ctx, "ORGANIZER")
	organizerID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:       organizerID,
		RoleID:   &orgRole.ID,
		Name:     "Organizer User",
		Email:    "contact@soundrhythm.com",
		Provider: "email",
		Status:   entity.UserStatusActive,
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := uc.UpsertProfile(ctx, uuid.New(), request.UpsertOrganizerProfileRequest{
			OrgName: "SoundRhythm",
		}, nil)
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})

	t.Run("invalid short orgName", func(t *testing.T) {
		_, err := uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
			OrgName: "AB",
		}, nil)
		if err == nil {
			t.Fatal("expected error for orgName < 3 chars")
		}
	})

	t.Run("invalid short phone", func(t *testing.T) {
		_, err := uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
			OrgName: "SoundRhythm",
			Phone:   "1234",
		}, nil)
		if err == nil {
			t.Fatal("expected error for phone < 8 chars")
		}
	})

	t.Run("success create organizer profile with logo", func(t *testing.T) {
		logo := &multipart.FileHeader{
			Filename: "logo.png",
			Size:     1024,
		}
		res, err := uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
			OrgName:     "SoundRhythm Live",
			Description: "International Music Promoter",
			Phone:       "081234567890",
			SocialMedia: "https://instagram.com/soundrhythm",
		}, logo)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.OrgName != "SoundRhythm Live" {
			t.Errorf("expected OrgName SoundRhythm Live, got %s", res.OrgName)
		}
		if res.Email != "contact@soundrhythm.com" {
			t.Errorf("expected Email contact@soundrhythm.com, got %s", res.Email)
		}
		if res.LogoURL == nil || *res.LogoURL == "" {
			t.Errorf("expected LogoURL to be set")
		}
	})

	t.Run("success update organizer profile preserves existing logo when no new logo uploaded", func(t *testing.T) {
		res, err := uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
			OrgName:     "SoundRhythm Updated",
			Description: "Updated Bio",
		}, nil)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.OrgName != "SoundRhythm Updated" {
			t.Errorf("expected OrgName SoundRhythm Updated, got %s", res.OrgName)
		}
		if res.LogoURL == nil || *res.LogoURL == "" {
			t.Errorf("expected existing LogoURL to be preserved")
		}
	})
}

func TestOrganizerUsecase_GetProfile(t *testing.T) {
	ctx := context.Background()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, inviteRepo, roleRepo, sessionRepo, storageService)

	orgRole, _ := roleRepo.FindByName(ctx, "ORGANIZER")
	organizerID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:       organizerID,
		RoleID:   &orgRole.ID,
		Name:     "Organizer User",
		Email:    "contact@soundrhythm.com",
		Provider: "email",
		Status:   entity.UserStatusActive,
	})

	t.Run("profile not found before setup", func(t *testing.T) {
		_, err := uc.GetProfile(ctx, organizerID)
		if err == nil {
			t.Fatal("expected error when profile does not exist yet")
		}
	})

	t.Run("success get profile", func(t *testing.T) {
		_, _ = uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
			OrgName:     "SoundRhythm Live",
			Description: "International Music Promoter",
			Phone:       "081234567890",
		}, nil)

		res, err := uc.GetProfile(ctx, organizerID)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.OrgName != "SoundRhythm Live" {
			t.Errorf("expected OrgName SoundRhythm Live, got %s", res.OrgName)
		}
		if res.Email != "contact@soundrhythm.com" {
			t.Errorf("expected Email contact@soundrhythm.com, got %s", res.Email)
		}
	})
}

func TestOrganizerUsecase_GetPublicProfile(t *testing.T) {
	ctx := context.Background()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, inviteRepo, roleRepo, sessionRepo, storageService)

	orgRole, _ := roleRepo.FindByName(ctx, "ORGANIZER")
	organizerID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:       organizerID,
		RoleID:   &orgRole.ID,
		Name:     "Organizer User",
		Email:    "public@soundrhythm.com",
		Provider: "email",
		Status:   entity.UserStatusActive,
	})

	created, _ := uc.UpsertProfile(ctx, organizerID, request.UpsertOrganizerProfileRequest{
		OrgName:     "SoundRhythm Public",
		Description: "Public Profile",
	}, nil)

	t.Run("not found for unknown id", func(t *testing.T) {
		_, err := uc.GetPublicProfile(ctx, uuid.New())
		if err == nil {
			t.Fatal("expected error for non-existent id")
		}
	})

	t.Run("found by profile ID", func(t *testing.T) {
		res, err := uc.GetPublicProfile(ctx, created.ID)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.OrgName != "SoundRhythm Public" {
			t.Errorf("expected OrgName SoundRhythm Public, got %s", res.OrgName)
		}
	})

	t.Run("found by organizer user ID", func(t *testing.T) {
		res, err := uc.GetPublicProfile(ctx, organizerID)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.ID != created.ID {
			t.Errorf("expected profile ID %s, got %s", created.ID, res.ID)
		}
	})

	t.Run("not found when organizer account is suspended", func(t *testing.T) {
		u, _ := userRepo.FindByID(ctx, organizerID)
		u.Status = entity.UserStatusSuspended
		_ = userRepo.Update(ctx, u)

		_, err := uc.GetPublicProfile(ctx, created.ID)
		if err == nil {
			t.Fatal("expected not found error when organizer is suspended")
		}
	})
}

func TestOrganizerUsecase_GetStaffList(t *testing.T) {
	ctx := context.Background()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, inviteRepo, roleRepo, sessionRepo, storageService)

	staffRole, _ := roleRepo.FindByName(ctx, "STAFF")
	organizerA := uuid.New()
	organizerB := uuid.New()

	// Staff 1 invited by Organizer A (ACTIVE)
	staff1ID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     staff1ID,
		RoleID: &staffRole.ID,
		Name:   "Alice Gate",
		Email:  "alice@staff.com",
		Status: entity.UserStatusActive,
	})
	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "alice@staff.com",
		RoleID:    staffRole.ID,
		InvitedBy: organizerA,
		TokenHash: "hash-alice",
		Status:    entity.InviteStatusAccepted,
	})

	// Staff 2 invited by Organizer A (SUSPENDED)
	staff2ID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     staff2ID,
		RoleID: &staffRole.ID,
		Name:   "Bob Scanner",
		Email:  "bob@staff.com",
		Status: entity.UserStatusSuspended,
	})
	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "bob@staff.com",
		RoleID:    staffRole.ID,
		InvitedBy: organizerA,
		TokenHash: "hash-bob",
		Status:    entity.InviteStatusAccepted,
	})

	// Staff 3 invited by Organizer B (should not appear for Organizer A)
	staff3ID := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     staff3ID,
		RoleID: &staffRole.ID,
		Name:   "Charlie Other",
		Email:  "charlie@staff.com",
		Status: entity.UserStatusActive,
	})
	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "charlie@staff.com",
		RoleID:    staffRole.ID,
		InvitedBy: organizerB,
		TokenHash: "hash-charlie",
		Status:    entity.InviteStatusAccepted,
	})

	t.Run("returns only staff invited by current organizer", func(t *testing.T) {
		res, err := uc.GetStaffList(ctx, organizerA, request.GetOrganizerStaffListRequest{
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(res.Items) != 2 {
			t.Fatalf("expected 2 staff for organizerA, got %d", len(res.Items))
		}
	})

	t.Run("filters staff by search and status", func(t *testing.T) {
		res, err := uc.GetStaffList(ctx, organizerA, request.GetOrganizerStaffListRequest{
			Page:     1,
			PageSize: 10,
			Search:   "bob",
			Status:   "SUSPENDED",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(res.Items) != 1 || res.Items[0].Email != "bob@staff.com" {
			t.Fatalf("expected only bob@staff.com, got %+v", res.Items)
		}
	})

	t.Run("returns empty list when organizer has no accepted invites", func(t *testing.T) {
		res, err := uc.GetStaffList(ctx, uuid.New(), request.GetOrganizerStaffListRequest{
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(res.Items) != 0 || res.Pagination.TotalRecords != 0 {
			t.Fatalf("expected 0 items, got %d", len(res.Items))
		}
	})
}

func TestOrganizerUsecase_UpdateStaffStatus(t *testing.T) {
	ctx := context.Background()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	inviteRepo := mock.NewInviteRepositoryMock()
	roleRepo := mock.NewRoleRepositoryMock()
	sessionRepo := mock.NewSessionRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, inviteRepo, roleRepo, sessionRepo, storageService)

	staffRole, _ := roleRepo.FindByName(ctx, "STAFF")
	userRole, _ := roleRepo.FindByName(ctx, "USER")
	organizerA := uuid.New()
	organizerB := uuid.New()

	// Staff belonging to Organizer A
	staffA := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     staffA,
		RoleID: &staffRole.ID,
		Name:   "Alice Gate",
		Email:  "alice@staff.com",
		Status: entity.UserStatusActive,
	})
	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "alice@staff.com",
		RoleID:    staffRole.ID,
		InvitedBy: organizerA,
		TokenHash: "hash-alice",
		Status:    entity.InviteStatusAccepted,
	})
	_ = sessionRepo.Create(ctx, &entity.Session{
		ID:        uuid.New(),
		UserID:    staffA,
		TokenHash: "session-hash-alice",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	// Staff belonging to Organizer B
	staffB := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     staffB,
		RoleID: &staffRole.ID,
		Name:   "Bob Other",
		Email:  "bob@staff.com",
		Status: entity.UserStatusActive,
	})
	_ = inviteRepo.Create(ctx, &entity.Invite{
		ID:        uuid.New(),
		Email:     "bob@staff.com",
		RoleID:    staffRole.ID,
		InvitedBy: organizerB,
		TokenHash: "hash-bob",
		Status:    entity.InviteStatusAccepted,
	})

	// Regular user
	regularUser := uuid.New()
	_ = userRepo.Create(ctx, &entity.User{
		ID:     regularUser,
		RoleID: &userRole.ID,
		Name:   "Buyer User",
		Email:  "buyer@user.com",
		Status: entity.UserStatusActive,
	})

	t.Run("success suspend own staff and revoke session", func(t *testing.T) {
		res, err := uc.UpdateStaffStatus(ctx, organizerA, staffA, request.UpdateUserStatusRequest{
			Status: "SUSPENDED",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Status != entity.UserStatusSuspended {
			t.Errorf("expected status SUSPENDED, got %s", res.Status)
		}

		sess, _ := sessionRepo.FindByTokenHash(ctx, "session-hash-alice")
		if sess != nil && sess.RevokedAt == nil {
			t.Error("expected staff session to be revoked upon suspension")
		}
	})

	t.Run("success reactivate own staff", func(t *testing.T) {
		res, err := uc.UpdateStaffStatus(ctx, organizerA, staffA, request.UpdateUserStatusRequest{
			Status: "ACTIVE",
		})
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if res.Status != entity.UserStatusActive {
			t.Errorf("expected status ACTIVE, got %s", res.Status)
		}
	})

	t.Run("forbidden when updating another organizer's staff", func(t *testing.T) {
		_, err := uc.UpdateStaffStatus(ctx, organizerA, staffB, request.UpdateUserStatusRequest{
			Status: "SUSPENDED",
		})
		if err == nil {
			t.Fatal("expected forbidden error when managing another organizer's staff")
		}
	})

	t.Run("forbidden when target user is not STAFF", func(t *testing.T) {
		_, err := uc.UpdateStaffStatus(ctx, organizerA, regularUser, request.UpdateUserStatusRequest{
			Status: "SUSPENDED",
		})
		if err == nil {
			t.Fatal("expected forbidden error when target is not STAFF")
		}
	})

	t.Run("bad request on invalid status value", func(t *testing.T) {
		_, err := uc.UpdateStaffStatus(ctx, organizerA, staffA, request.UpdateUserStatusRequest{
			Status: "BANNED",
		})
		if err == nil {
			t.Fatal("expected bad request error for invalid status")
		}
	})
}

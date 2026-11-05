package usecase_test

import (
	"context"
	"mime/multipart"
	"testing"

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
	roleRepo := mock.NewRoleRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, storageService)

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
	roleRepo := mock.NewRoleRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, storageService)

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
	roleRepo := mock.NewRoleRepositoryMock()
	storageService := &mockStorageService{}

	uc := usecase.NewOrganizerUsecase(organizerRepo, userRepo, storageService)

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

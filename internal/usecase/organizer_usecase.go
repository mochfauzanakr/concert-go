package usecase

import (
	"context"
	"mime/multipart"
	"strings"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

type OrganizerUsecase struct {
	organizerRepo  repository.OrganizerRepository
	userRepo       repository.UserRepository
	storageService util.StorageService
}

func NewOrganizerUsecase(
	organizerRepo repository.OrganizerRepository,
	userRepo repository.UserRepository,
	storageService util.StorageService,
) *OrganizerUsecase {
	return &OrganizerUsecase{
		organizerRepo:  organizerRepo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

func (u *OrganizerUsecase) UpsertProfile(
	ctx context.Context,
	userID uuid.UUID,
	req request.UpsertOrganizerProfileRequest,
	logoHeader *multipart.FileHeader,
) (*response.OrganizerProfileResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, exception.NotFound("organizer user not found")
	}

	orgName := strings.TrimSpace(req.OrgName)
	if len(orgName) < 3 || len(orgName) > 100 {
		return nil, exception.BadRequest("orgName must be between 3 and 100 characters")
	}

	var phonePtr *string
	if phone := strings.TrimSpace(req.Phone); phone != "" {
		if len(phone) < 8 || len(phone) > 20 {
			return nil, exception.BadRequest("phone must be between 8 and 20 characters")
		}
		phonePtr = &phone
	}

	var descPtr *string
	if desc := strings.TrimSpace(req.Description); desc != "" {
		descPtr = &desc
	}

	var socialPtr *string
	if social := strings.TrimSpace(req.SocialMedia); social != "" {
		socialPtr = &social
	}

	var logoURLPtr *string
	if logoHeader != nil {
		if u.storageService == nil {
			return nil, exception.Internal("storage service not configured", nil)
		}
		uploadedURL, err := u.storageService.Upload(ctx, logoHeader, "logos")
		if err != nil {
			return nil, err
		}
		logoURLPtr = &uploadedURL
	}

	profile := &entity.OrganizerProfile{
		UserID:      userID,
		OrgName:     orgName,
		LogoURL:     logoURLPtr,
		Description: descPtr,
		Phone:       phonePtr,
		SocialMedia: socialPtr,
	}

	if err := u.organizerRepo.Upsert(ctx, profile); err != nil {
		return nil, err
	}

	return toOrganizerProfileResponse(profile, user.Email), nil
}

func (u *OrganizerUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*response.OrganizerProfileResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, exception.NotFound("organizer user not found")
	}

	profile, err := u.organizerRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, exception.NotFound("organizer profile not found")
	}

	return toOrganizerProfileResponse(profile, user.Email), nil
}

func (u *OrganizerUsecase) GetPublicProfile(ctx context.Context, id uuid.UUID) (*response.OrganizerProfileResponse, error) {
	profile, err := u.organizerRepo.FindByIDOrUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, exception.NotFound("organizer profile not found")
	}

	user, err := u.userRepo.FindByID(ctx, profile.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status == entity.UserStatusSuspended {
		return nil, exception.NotFound("organizer profile not found")
	}

	return toOrganizerProfileResponse(profile, user.Email), nil
}

func toOrganizerProfileResponse(profile *entity.OrganizerProfile, email string) *response.OrganizerProfileResponse {
	return &response.OrganizerProfileResponse{
		ID:          profile.ID,
		UserID:      profile.UserID,
		OrgName:     profile.OrgName,
		Email:       email,
		LogoURL:     profile.LogoURL,
		Description: profile.Description,
		Phone:       profile.Phone,
		SocialMedia: profile.SocialMedia,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
	}
}

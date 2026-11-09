package usecase

import (
	"context"
	"mime/multipart"
	"strings"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
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
	inviteRepo     repository.InviteRepository
	roleRepo       repository.RoleRepository
	sessionRepo    repository.SessionRepository
	storageService util.StorageService
}

func NewOrganizerUsecase(
	organizerRepo repository.OrganizerRepository,
	userRepo repository.UserRepository,
	inviteRepo repository.InviteRepository,
	roleRepo repository.RoleRepository,
	sessionRepo repository.SessionRepository,
	storageService util.StorageService,
) *OrganizerUsecase {
	return &OrganizerUsecase{
		organizerRepo:  organizerRepo,
		userRepo:       userRepo,
		inviteRepo:     inviteRepo,
		roleRepo:       roleRepo,
		sessionRepo:    sessionRepo,
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

func (u *OrganizerUsecase) GetStaffList(
	ctx context.Context,
	organizerID uuid.UUID,
	req request.GetOrganizerStaffListRequest,
) (*response.PageResponse[response.UserListItemResponse], error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 20 {
		pageSize = 20
	}

	emails, err := u.inviteRepo.FindAcceptedEmailsByInviter(ctx, organizerID)
	if err != nil {
		return nil, err
	}
	if len(emails) == 0 {
		return &response.PageResponse[response.UserListItemResponse]{
			Items:      []response.UserListItemResponse{},
			Pagination: model.BuildPagination(0, page, pageSize),
		}, nil
	}

	var staffRoleID *int
	if u.roleRepo != nil {
		if r, err := u.roleRepo.FindByName(ctx, "STAFF"); err == nil && r != nil {
			staffRoleID = &r.ID
		}
	}

	filter := repository.UserFilter{
		Search: strings.TrimSpace(req.Search),
		RoleID: staffRoleID,
		Status: strings.ToUpper(strings.TrimSpace(req.Status)),
		Emails: emails,
		Limit:  pageSize,
		Offset: model.CalculateOffset(page, pageSize),
	}

	users, total, err := u.userRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]response.UserListItemResponse, 0, len(users))
	for _, user := range users {
		items = append(items, response.UserListItemResponse{
			ID:        user.ID,
			RoleID:    user.RoleID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		})
	}

	return &response.PageResponse[response.UserListItemResponse]{
		Items:      items,
		Pagination: model.BuildPagination(total, page, pageSize),
	}, nil
}

func (u *OrganizerUsecase) UpdateStaffStatus(
	ctx context.Context,
	organizerID uuid.UUID,
	staffID uuid.UUID,
	req request.UpdateUserStatusRequest,
) (*response.UserDetailResponse, error) {
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != entity.UserStatusActive && status != entity.UserStatusSuspended {
		return nil, exception.BadRequest("status must be ACTIVE or SUSPENDED")
	}

	targetUser, err := u.userRepo.FindByID(ctx, staffID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, exception.NotFound("staff user not found")
	}

	targetRoleName := "UNKNOWN"
	if targetUser.RoleID != nil && u.roleRepo != nil {
		if r, err := u.roleRepo.FindByID(ctx, *targetUser.RoleID); err == nil && r != nil {
			targetRoleName = r.Name
		}
	}
	if targetRoleName != "STAFF" {
		return nil, exception.Forbidden("target user is not a staff member")
	}

	emails, err := u.inviteRepo.FindAcceptedEmailsByInviter(ctx, organizerID)
	if err != nil {
		return nil, err
	}

	owned := false
	for _, email := range emails {
		if strings.EqualFold(email, targetUser.Email) {
			owned = true
			break
		}
	}
	if !owned {
		return nil, exception.Forbidden("you do not have permission to manage this staff")
	}

	targetUser.Status = status
	if err := u.userRepo.Update(ctx, targetUser); err != nil {
		return nil, err
	}

	if status == entity.UserStatusSuspended && u.sessionRepo != nil {
		_ = u.sessionRepo.RevokeByUserID(ctx, staffID)
	}

	return &response.UserDetailResponse{
		ID:        targetUser.ID,
		RoleID:    targetUser.RoleID,
		Name:      targetUser.Name,
		Email:     targetUser.Email,
		AvatarURL: targetUser.AvatarURL,
		Provider:  targetUser.Provider,
		Status:    targetUser.Status,
		CreatedAt: targetUser.CreatedAt,
	}, nil
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

package usecase

import (
	"context"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"concert-go/internal/config"
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
	emailSender    util.EmailSender
	storageService util.StorageService
	cfg            *config.Config
}

func NewOrganizerUsecase(
	organizerRepo repository.OrganizerRepository,
	userRepo repository.UserRepository,
	inviteRepo repository.InviteRepository,
	roleRepo repository.RoleRepository,
	sessionRepo repository.SessionRepository,
	emailSender util.EmailSender,
	storageService util.StorageService,
	cfg *config.Config,
) *OrganizerUsecase {
	return &OrganizerUsecase{
		organizerRepo:  organizerRepo,
		userRepo:       userRepo,
		inviteRepo:     inviteRepo,
		roleRepo:       roleRepo,
		sessionRepo:    sessionRepo,
		emailSender:    emailSender,
		storageService: storageService,
		cfg:            cfg,
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

	var staffRoleID *int
	if u.roleRepo != nil {
		if r, err := u.roleRepo.FindByName(ctx, "STAFF"); err == nil && r != nil {
			staffRoleID = &r.ID
		}
	}

	filter := repository.UserFilter{
		Search:      strings.TrimSpace(req.Search),
		RoleID:      staffRoleID,
		Status:      strings.ToUpper(strings.TrimSpace(req.Status)),
		OrganizerID: &organizerID,
		Limit:       pageSize,
		Offset:      model.CalculateOffset(page, pageSize),
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

	owned, err := u.organizerRepo.IsStaffOfOrganizer(ctx, organizerID, staffID)
	if err != nil {
		return nil, err
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

func (u *OrganizerUsecase) ApplyForPartnership(
	ctx context.Context,
	req request.ApplyOrganizerRequest,
) (*response.OrganizerApplicationResponse, error) {
	picName := strings.TrimSpace(req.PicName)
	if len(picName) < 2 || len(picName) > 100 {
		return nil, exception.BadRequest("picName must be between 2 and 100 characters")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := util.ValidateEmailTLD(email); err != nil {
		return nil, exception.BadRequest(err.Error())
	}

	phone := strings.TrimSpace(req.Phone)
	if len(phone) < 8 || len(phone) > 20 {
		return nil, exception.BadRequest("phone must be between 8 and 20 characters")
	}

	orgName := strings.TrimSpace(req.OrgName)
	if len(orgName) < 3 || len(orgName) > 100 {
		return nil, exception.BadRequest("orgName must be between 3 and 100 characters")
	}

	description := strings.TrimSpace(req.Description)
	if len(description) < 10 {
		return nil, exception.BadRequest("description must be at least 10 characters")
	}

	var socialPtr *string
	if social := strings.TrimSpace(req.SocialMedia); social != "" {
		socialPtr = &social
	}

	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, exception.Conflict("email is already registered in the system")
	}

	latestApp, err := u.organizerRepo.FindLatestApplicationByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if latestApp != nil && latestApp.Status == entity.ApplicationStatusPending {
		return nil, exception.Conflict("a pending application already exists for this email")
	}

	app := &entity.OrganizerApplication{
		ID:          uuid.New(),
		PicName:     picName,
		Email:       email,
		Phone:       phone,
		OrgName:     orgName,
		SocialMedia: socialPtr,
		Description: description,
		Status:      entity.ApplicationStatusPending,
		CreatedAt:   time.Now(),
	}

	if err := u.organizerRepo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}

	return toOrganizerApplicationResponse(app), nil
}

func (u *OrganizerUsecase) GetApplications(
	ctx context.Context,
	req request.GetOrganizerApplicationListRequest,
) (*response.PageResponse[response.OrganizerApplicationResponse], error) {
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

	filter := repository.OrganizerApplicationFilter{
		Search: strings.TrimSpace(req.Search),
		Status: strings.ToUpper(strings.TrimSpace(req.Status)),
		Limit:  pageSize,
		Offset: model.CalculateOffset(page, pageSize),
	}

	apps, total, err := u.organizerRepo.FindAllApplications(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]response.OrganizerApplicationResponse, 0, len(apps))
	for _, app := range apps {
		items = append(items, *toOrganizerApplicationResponse(app))
	}

	return &response.PageResponse[response.OrganizerApplicationResponse]{
		Items:      items,
		Pagination: model.BuildPagination(total, page, pageSize),
	}, nil
}

func (u *OrganizerUsecase) ReviewApplication(
	ctx context.Context,
	reviewerID uuid.UUID,
	appID uuid.UUID,
	req request.ReviewOrganizerApplicationRequest,
) (*response.OrganizerApplicationResponse, error) {
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != entity.ApplicationStatusApproved && status != entity.ApplicationStatusRejected {
		return nil, exception.BadRequest("status must be APPROVED or REJECTED")
	}

	app, err := u.organizerRepo.FindApplicationByID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, exception.NotFound("organizer application not found")
	}
	if app.Status != entity.ApplicationStatusPending {
		return nil, exception.Conflict("application has already been reviewed")
	}

	now := time.Now()
	app.Status = status
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now

	if status == entity.ApplicationStatusRejected {
		if reason := strings.TrimSpace(req.RejectionReason); reason != "" {
			app.RejectionReason = &reason
		}
	} else {
		existingUser, err := u.userRepo.FindByEmail(ctx, app.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, exception.Conflict("email is already registered in the system")
		}

		pendingInvite, err := u.inviteRepo.FindPendingByEmail(ctx, app.Email)
		if err != nil {
			return nil, err
		}
		if pendingInvite != nil {
			return nil, exception.Conflict("an active invitation already exists for this email")
		}

		orgRole, err := u.roleRepo.FindByName(ctx, "ORGANIZER")
		if err != nil || orgRole == nil {
			return nil, exception.Internal("ORGANIZER role not found", err)
		}

		rawToken := uuid.New().String() + "-" + uuid.New().String()
		tokenHash := util.HashToken(rawToken)
		invite := &entity.Invite{
			ID:        uuid.New(),
			Email:     app.Email,
			RoleID:    orgRole.ID,
			InvitedBy: reviewerID,
			TokenHash: tokenHash,
			Status:    entity.InviteStatusPending,
			ExpiresAt: now.Add(24 * time.Hour),
		}

		baseURL := "http://localhost:3100"
		if u.cfg != nil && strings.TrimSpace(u.cfg.FrontendURL) != "" {
			baseURL = strings.TrimRight(strings.TrimSpace(u.cfg.FrontendURL), "/")
		}
		inviteURL := baseURL + "/accept-invite?token=" + rawToken
		log.Printf("[DEV] Organizer Application Invite URL (%s): %s | Token: %s", app.Email, inviteURL, rawToken) // ponytail: stdout log for local testing; remove or gate by env before prod
		if u.emailSender != nil {
			if err := u.emailSender.SendInviteLink(app.Email, orgRole.Name, inviteURL); err != nil {
				return nil, exception.Internal("failed to send invitation email", err)
			}
		}

		if err := u.inviteRepo.Create(ctx, invite); err != nil {
			return nil, err
		}
	}

	if err := u.organizerRepo.UpdateApplicationReview(ctx, app); err != nil {
		return nil, err
	}

	return toOrganizerApplicationResponse(app), nil
}

func toOrganizerApplicationResponse(app *entity.OrganizerApplication) *response.OrganizerApplicationResponse {
	return &response.OrganizerApplicationResponse{
		ID:              app.ID,
		PicName:         app.PicName,
		Email:           app.Email,
		Phone:           app.Phone,
		OrgName:         app.OrgName,
		SocialMedia:     app.SocialMedia,
		Description:     app.Description,
		Status:          app.Status,
		RejectionReason: app.RejectionReason,
		ReviewedBy:      app.ReviewedBy,
		ReviewedAt:      app.ReviewedAt,
		CreatedAt:       app.CreatedAt,
	}
}

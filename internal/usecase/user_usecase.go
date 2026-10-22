package usecase

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

type UserUsecase struct {
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	inviteRepo  repository.InviteRepository
	emailSender util.EmailSender
	cfg         *config.Config
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	inviteRepo repository.InviteRepository,
	emailSender util.EmailSender,
	cfg *config.Config,
) *UserUsecase {
	return &UserUsecase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		inviteRepo:  inviteRepo,
		emailSender: emailSender,
		cfg:         cfg,
	}
}

func (u *UserUsecase) InviteUser(ctx context.Context, inviterID uuid.UUID, inviterRoleID *int, req request.InviteUserRequest) (*response.InviteResponse, error) {
	if err := util.ValidateEmailTLD(req.Email); err != nil {
		return nil, exception.BadRequest(err.Error())
	}

	if inviterRoleID == nil {
		return nil, exception.Forbidden("inviter has no assigned role")
	}

	inviterRole, err := u.roleRepo.FindByID(ctx, *inviterRoleID)
	if err != nil || inviterRole == nil {
		return nil, exception.Forbidden("inviter role not found")
	}

	targetRoleName := strings.ToUpper(strings.TrimSpace(req.Role))
	targetRole, err := u.roleRepo.FindByName(ctx, targetRoleName)
	if err != nil || targetRole == nil {
		return nil, exception.BadRequest("invalid role specified")
	}

	switch strings.ToUpper(inviterRole.Name) {
	case "SUPER_ADMIN":
		if targetRoleName != "ADMIN" && targetRoleName != "ORGANIZER" {
			return nil, exception.Forbidden("SUPER_ADMIN can only invite ADMIN or ORGANIZER")
		}
	case "ADMIN":
		if targetRoleName != "ORGANIZER" {
			return nil, exception.Forbidden("ADMIN can only invite ORGANIZER")
		}
	case "ORGANIZER":
		if targetRoleName != "STAFF" {
			return nil, exception.Forbidden("ORGANIZER can only invite STAFF")
		}
	default:
		return nil, exception.Forbidden("you do not have permission to invite users")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, exception.Conflict("email is already registered in the system")
	}

	pendingInvite, err := u.inviteRepo.FindPendingByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if pendingInvite != nil {
		return nil, exception.Conflict("an active invitation already exists for this email")
	}

	rawToken := uuid.New().String() + "-" + uuid.New().String()
	tokenHash := util.HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	invite := &entity.Invite{
		ID:        uuid.New(),
		Email:     email,
		RoleID:    targetRole.ID,
		InvitedBy: inviterID,
		TokenHash: tokenHash,
		Status:    entity.InviteStatusPending,
		ExpiresAt: expiresAt,
	}

	if err := u.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	inviteURL := "/accept-invite?token=" + rawToken
	if u.emailSender != nil {
		if err := u.emailSender.SendInviteLink(email, targetRole.Name, inviteURL); err != nil {
			return nil, exception.Internal("failed to send invitation email", err)
		}
	}

	return &response.InviteResponse{
		ID:        invite.ID,
		Email:     invite.Email,
		Role:      targetRole.Name,
		Status:    invite.Status,
		ExpiresAt: invite.ExpiresAt,
	}, nil
}

func (u *UserUsecase) AcceptInvite(ctx context.Context, req request.AcceptInviteRequest) (*response.UserResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, exception.BadRequest("password and confirm password do not match")
	}
	if err := util.ValidatePassword(req.Password); err != nil {
		return nil, exception.BadRequest(err.Error())
	}

	tokenHash := util.HashToken(req.Token)
	invite, err := u.inviteRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if invite == nil || invite.Status != entity.InviteStatusPending || time.Now().After(invite.ExpiresAt) {
		return nil, exception.BadRequest("invalid or expired invitation token")
	}

	existingUser, err := u.userRepo.FindByEmail(ctx, invite.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, exception.Conflict("email is already registered")
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		RoleID:       &invite.RoleID,
		Name:         strings.TrimSpace(req.Name),
		Email:        invite.Email,
		PasswordHash: &hashedPassword,
		Provider:     "email",
		Status:       entity.UserStatusActive,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	now := time.Now()
	_ = u.inviteRepo.UpdateStatus(ctx, invite.ID, entity.InviteStatusAccepted, &now)

	return &response.UserResponse{
		ID:        user.ID,
		RoleID:    user.RoleID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (u *UserUsecase) ValidateInvite(ctx context.Context, token string) (*response.InviteResponse, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, exception.BadRequest("invitation token is required")
	}

	tokenHash := util.HashToken(token)
	invite, err := u.inviteRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if invite == nil || invite.Status != entity.InviteStatusPending || time.Now().After(invite.ExpiresAt) {
		return nil, exception.BadRequest("invalid or expired invitation token")
	}

	roleName := "UNKNOWN"
	if u.roleRepo != nil {
		role, err := u.roleRepo.FindByID(ctx, invite.RoleID)
		if err == nil && role != nil {
			roleName = role.Name
		}
	}

	return &response.InviteResponse{
		ID:        invite.ID,
		Email:     invite.Email,
		Role:      roleName,
		Status:    invite.Status,
		ExpiresAt: invite.ExpiresAt,
	}, nil
}

func (u *UserUsecase) GetAllUsers(ctx context.Context, req request.GetUserListRequest) (*response.PageResponse[response.UserListItemResponse], error) {
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

	search := strings.TrimSpace(req.Search)
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	offset := model.CalculateOffset(page, pageSize)

	filter := repository.UserFilter{
		Search: search,
		RoleID: req.RoleID,
		Status: status,
		Limit:  pageSize,
		Offset: offset,
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

	pagination := model.BuildPagination(total, page, pageSize)

	return &response.PageResponse[response.UserListItemResponse]{
		Items:      items,
		Pagination: pagination,
	}, nil
}

func (u *UserUsecase) UpdateUserStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRoleID *int,
	targetUserID uuid.UUID,
	req request.UpdateUserStatusRequest,
) (*response.UserDetailResponse, error) {
	if actorID == targetUserID {
		return nil, exception.BadRequest("cannot change your own account status")
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != entity.UserStatusActive && status != entity.UserStatusSuspended {
		return nil, exception.BadRequest("status must be ACTIVE or SUSPENDED")
	}

	actorRoleName := ""
	if actorRoleID != nil && u.roleRepo != nil {
		r, err := u.roleRepo.FindByID(ctx, *actorRoleID)
		if err == nil && r != nil {
			actorRoleName = r.Name
		}
	}

	targetUser, err := u.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, exception.NotFound("user not found")
	}

	targetRoleName := "UNKNOWN"
	if targetUser.RoleID != nil && u.roleRepo != nil {
		r, err := u.roleRepo.FindByID(ctx, *targetUser.RoleID)
		if err == nil && r != nil {
			targetRoleName = r.Name
		}
	}

	if targetRoleName == "SUPER_ADMIN" {
		return nil, exception.Forbidden("cannot change status of SUPER_ADMIN")
	}

	if actorRoleName == "ADMIN" && (targetRoleName == "ADMIN" || targetRoleName == "SUPER_ADMIN") {
		return nil, exception.Forbidden("admins cannot change status of other admins or super admins")
	}

	targetUser.Status = status
	if err := u.userRepo.Update(ctx, targetUser); err != nil {
		return nil, err
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


func (u *UserUsecase) AdminUpdateUser(
	ctx context.Context,
	actorID uuid.UUID,
	actorRoleID *int,
	targetUserID uuid.UUID,
	req request.AdminUpdateUserRequest,
) (*response.UserDetailResponse, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 || len(name) > 100 {
		return nil, exception.BadRequest("name must be between 2 and 100 characters")
	}

	actorRoleName := ""
	if actorRoleID != nil && u.roleRepo != nil {
		r, err := u.roleRepo.FindByID(ctx, *actorRoleID)
		if err == nil && r != nil {
			actorRoleName = r.Name
		}
	}

	targetUser, err := u.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, exception.NotFound("user not found")
	}

	targetRoleName := "UNKNOWN"
	if targetUser.RoleID != nil && u.roleRepo != nil {
		r, err := u.roleRepo.FindByID(ctx, *targetUser.RoleID)
		if err == nil && r != nil {
			targetRoleName = r.Name
		}
	}

	newRole, err := u.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil || newRole == nil {
		return nil, exception.BadRequest("role not found")
	}

	// 1. Cannot change your own role
	if actorID == targetUserID && actorRoleID != nil && *actorRoleID != req.RoleID {
		return nil, exception.Forbidden("cannot change your own role")
	}

	// 2. ADMIN cannot edit SUPER_ADMIN or other ADMIN
	if actorRoleName == "ADMIN" && (targetRoleName == "SUPER_ADMIN" || targetRoleName == "ADMIN") {
		return nil, exception.Forbidden("admins cannot edit other admins or super admins")
	}

	// 3. Only SUPER_ADMIN can assign SUPER_ADMIN role
	if newRole.Name == "SUPER_ADMIN" && actorRoleName != "SUPER_ADMIN" {
		return nil, exception.Forbidden("only super admin can assign super admin role")
	}
	if targetRoleName == "SUPER_ADMIN" && actorRoleName != "SUPER_ADMIN" {
		return nil, exception.Forbidden("cannot edit super admin")
	}

	targetUser.Name = name
	targetUser.RoleID = &req.RoleID

	if err := u.userRepo.Update(ctx, targetUser); err != nil {
		return nil, err
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

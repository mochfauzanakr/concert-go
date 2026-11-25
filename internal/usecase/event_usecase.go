package usecase

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/util"

	"github.com/google/uuid"
)

type EventUsecase struct {
	eventRepo      repository.EventRepository
	organizerRepo  repository.OrganizerRepository
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	storageService util.StorageService
}

func NewEventUsecase(
	eventRepo repository.EventRepository,
	organizerRepo repository.OrganizerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	storageService util.StorageService,
) *EventUsecase {
	return &EventUsecase{
		eventRepo:      eventRepo,
		organizerRepo:  organizerRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		storageService: storageService,
	}
}

type parsedEventInput struct {
	name          string
	description   string
	category      string
	city          string
	location      string
	startTime     time.Time
	endTime       time.Time
	inventoryMode string
}

func validateAndParseEventInput(req request.UpsertEventRequest) (*parsedEventInput, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) < 3 || len(name) > 150 {
		return nil, exception.BadRequest("name must be between 3 and 150 characters")
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return nil, exception.BadRequest("description is required")
	}
	category := strings.TrimSpace(req.Category)
	if category == "" {
		return nil, exception.BadRequest("category is required")
	}
	city := strings.TrimSpace(req.City)
	if city == "" {
		return nil, exception.BadRequest("city is required")
	}
	location := strings.TrimSpace(req.Location)
	if location == "" {
		return nil, exception.BadRequest("location is required")
	}

	startTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.StartTime))
	if err != nil {
		return nil, exception.BadRequest("startTime must be a valid ISO-8601 timestamp")
	}
	endTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EndTime))
	if err != nil {
		return nil, exception.BadRequest("endTime must be a valid ISO-8601 timestamp")
	}
	if !endTime.After(startTime) {
		return nil, exception.BadRequest("endTime must be after startTime")
	}

	inventoryMode := strings.ToUpper(strings.TrimSpace(req.InventoryMode))
	if inventoryMode != entity.InventoryModeOnlineOnly && inventoryMode != entity.InventoryModeDualChannel {
		return nil, exception.BadRequest("inventoryMode must be ONLINE_ONLY or DUAL_CHANNEL")
	}

	return &parsedEventInput{
		name:          name,
		description:   description,
		category:      category,
		city:          city,
		location:      location,
		startTime:     startTime,
		endTime:       endTime,
		inventoryMode: inventoryMode,
	}, nil
}

func (u *EventUsecase) uploadOptionalImage(ctx context.Context, header *multipart.FileHeader, folder string) (*string, error) {
	if header == nil {
		return nil, nil
	}
	if u.storageService == nil {
		return nil, exception.Internal("storage service not configured", nil)
	}
	uploadedURL, err := u.storageService.Upload(ctx, header, folder)
	if err != nil {
		return nil, err
	}
	return &uploadedURL, nil
}

func buildEventListResponse(events []*entity.Event, total, page, pageSize int) *response.PageResponse[response.EventListItemResponse] {
	items := make([]response.EventListItemResponse, 0, len(events))
	for _, ev := range events {
		items = append(items, response.EventListItemResponse{
			ID:            ev.ID,
			OrganizerID:   ev.OrganizerID,
			Name:          ev.Name,
			BannerURL:     ev.BannerURL,
			Category:      ev.Category,
			City:          ev.City,
			Location:      ev.Location,
			StartTime:     ev.StartTime,
			EndTime:       ev.EndTime,
			Status:        ev.Status,
			InventoryMode: ev.InventoryMode,
			CreatedAt:     ev.CreatedAt,
		})
	}
	return &response.PageResponse[response.EventListItemResponse]{
		Items:      items,
		Pagination: model.BuildPagination(total, page, pageSize),
	}
}

func (u *EventUsecase) buildEventDetail(ctx context.Context, ev *entity.Event) (*response.EventDetailResponse, error) {
	var leadOrg *response.EventOrganizerSummaryResponse
	if u.organizerRepo != nil {
		profile, err := u.organizerRepo.FindByUserID(ctx, ev.OrganizerID)
		if err != nil {
			return nil, err
		}
		if profile != nil {
			profileID := profile.ID
			leadOrg = &response.EventOrganizerSummaryResponse{
				UserID:    ev.OrganizerID,
				ProfileID: &profileID,
				OrgName:   profile.OrgName,
				LogoURL:   profile.LogoURL,
			}
		}
	}
	if leadOrg == nil && u.userRepo != nil {
		user, err := u.userRepo.FindByID(ctx, ev.OrganizerID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			leadOrg = &response.EventOrganizerSummaryResponse{
				UserID:  ev.OrganizerID,
				OrgName: user.Name,
			}
		}
	}

	rawCoOrgs, err := u.eventRepo.FindCoOrganizersByEventID(ctx, ev.ID)
	if err != nil {
		return nil, err
	}
	coOrgs := make([]response.EventCoOrganizerResponse, 0, len(rawCoOrgs))
	for _, co := range rawCoOrgs {
		item := response.EventCoOrganizerResponse{
			UserID:     co.UserID,
			Name:       co.Name,
			OrgName:    co.OrgName,
			LogoURL:    co.LogoURL,
			AssignedAt: co.AssignedAt,
		}
		if item.Name == "" && u.userRepo != nil {
			if user, err := u.userRepo.FindByID(ctx, co.UserID); err == nil && user != nil {
				item.Name = user.Name
			}
		}
		if item.OrgName == "" && u.organizerRepo != nil {
			if profile, err := u.organizerRepo.FindByUserID(ctx, co.UserID); err == nil && profile != nil {
				item.OrgName = profile.OrgName
				item.LogoURL = profile.LogoURL
			}
		}
		if item.OrgName == "" {
			item.OrgName = item.Name
		}
		coOrgs = append(coOrgs, item)
	}

	rawArtists, err := u.eventRepo.FindArtistsByEventID(ctx, ev.ID)
	if err != nil {
		return nil, err
	}
	artists := make([]response.EventArtistResponse, 0, len(rawArtists))
	for _, a := range rawArtists {
		artists = append(artists, response.EventArtistResponse{
			ID:        a.ID,
			Name:      a.Name,
			Bio:       a.Bio,
			AvatarURL: a.AvatarURL,
		})
	}

	return &response.EventDetailResponse{
		ID:            ev.ID,
		Name:          ev.Name,
		Description:   ev.Description,
		BannerURL:     ev.BannerURL,
		Category:      ev.Category,
		City:          ev.City,
		Location:      ev.Location,
		StartTime:     ev.StartTime,
		EndTime:       ev.EndTime,
		Status:        ev.Status,
		InventoryMode: ev.InventoryMode,
		LeadOrganizer: leadOrg,
		CoOrganizers:  coOrgs,
		Artists:       artists,
		// ponytail: ticket/inventory summary defaults to zero until FD-5.0 tables are implemented
		TicketSummary: response.EventTicketSummaryResponse{
			TotalAvailable: 0,
			LowestPrice:    nil,
			SoldOut:        false,
		},
		CreatedAt: ev.CreatedAt,
		UpdatedAt: ev.UpdatedAt,
	}, nil
}

// GetPublicEvents lists PUBLISHED events for the public catalog (F-4.1).
func (u *EventUsecase) GetPublicEvents(ctx context.Context, req request.GetPublicEventListRequest) (*response.PageResponse[response.EventListItemResponse], error) {
	req.Normalize(20)

	filter := repository.EventFilter{
		Search:    req.Search,
		City:      strings.TrimSpace(req.City),
		Category:  strings.TrimSpace(req.Category),
		Status:    entity.EventStatusPublished,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Limit:     req.PageSize,
		Offset:    model.CalculateOffset(req.Page, req.PageSize),
	}

	events, total, err := u.eventRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildEventListResponse(events, total, req.Page, req.PageSize), nil
}

// GetPublicEventDetail returns a PUBLISHED event's full detail (F-4.1).
func (u *EventUsecase) GetPublicEventDetail(ctx context.Context, eventID uuid.UUID) (*response.EventDetailResponse, error) {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil || ev.Status != entity.EventStatusPublished {
		return nil, exception.NotFound("event not found")
	}
	return u.buildEventDetail(ctx, ev)
}

// GetOrganizerEvents lists all events managed by the organizer as lead or co-organizer (F-4.2).
func (u *EventUsecase) GetOrganizerEvents(ctx context.Context, organizerID uuid.UUID, req request.GetOrganizerEventListRequest) (*response.PageResponse[response.EventListItemResponse], error) {
	req.Normalize(20)

	filter := repository.EventFilter{
		Search:          req.Search,
		City:            strings.TrimSpace(req.City),
		Category:        strings.TrimSpace(req.Category),
		Status:          strings.ToUpper(strings.TrimSpace(req.Status)),
		ManagedByUserID: &organizerID,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		SortBy:          req.SortBy,
		SortOrder:       req.SortOrder,
		Limit:           req.PageSize,
		Offset:          model.CalculateOffset(req.Page, req.PageSize),
	}

	events, total, err := u.eventRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildEventListResponse(events, total, req.Page, req.PageSize), nil
}

// CreateEvent creates a new DRAFT event for an organizer with a completed profile (F-4.2).
func (u *EventUsecase) CreateEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	req request.UpsertEventRequest,
	bannerHeader *multipart.FileHeader,
) (*response.EventDetailResponse, error) {
	if u.organizerRepo != nil {
		profile, err := u.organizerRepo.FindByUserID(ctx, organizerID)
		if err != nil {
			return nil, err
		}
		if profile == nil {
			return nil, exception.UnprocessableEntity("organizer profile must be completed before creating an event")
		}
	}

	parsed, err := validateAndParseEventInput(req)
	if err != nil {
		return nil, err
	}

	bannerURL, err := u.uploadOptionalImage(ctx, bannerHeader, "banners")
	if err != nil {
		return nil, err
	}

	ev := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   organizerID,
		Name:          parsed.name,
		Description:   parsed.description,
		BannerURL:     bannerURL,
		Category:      parsed.category,
		City:          parsed.city,
		Location:      parsed.location,
		StartTime:     parsed.startTime,
		EndTime:       parsed.endTime,
		Status:        entity.EventStatusDraft,
		InventoryMode: parsed.inventoryMode,
	}

	if err := u.eventRepo.Create(ctx, ev); err != nil {
		return nil, err
	}
	return u.buildEventDetail(ctx, ev)
}

// UpdateEvent updates event metadata and optional banner for the Lead Organizer (F-4.2).
func (u *EventUsecase) UpdateEvent(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	req request.UpsertEventRequest,
	bannerHeader *multipart.FileHeader,
) (*response.EventDetailResponse, error) {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, exception.NotFound("event not found")
	}
	if ev.OrganizerID != organizerID {
		return nil, exception.Forbidden("only the lead organizer can update this event")
	}

	parsed, err := validateAndParseEventInput(req)
	if err != nil {
		return nil, err
	}

	bannerURL, err := u.uploadOptionalImage(ctx, bannerHeader, "banners")
	if err != nil {
		return nil, err
	}

	ev.Name = parsed.name
	ev.Description = parsed.description
	ev.Category = parsed.category
	ev.City = parsed.city
	ev.Location = parsed.location
	ev.StartTime = parsed.startTime
	ev.EndTime = parsed.endTime
	ev.InventoryMode = parsed.inventoryMode
	if bannerURL != nil {
		ev.BannerURL = bannerURL
	}

	if err := u.eventRepo.Update(ctx, ev); err != nil {
		return nil, err
	}
	return u.buildEventDetail(ctx, ev)
}

// UpdateEventStatus updates the lifecycle status of an event by its Lead Organizer (F-4.2).
func (u *EventUsecase) UpdateEventStatus(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	req request.UpdateEventStatusRequest,
) (*response.EventDetailResponse, error) {
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	switch status {
	case entity.EventStatusDraft, entity.EventStatusPublished, entity.EventStatusCompleted, entity.EventStatusCancelled:
	default:
		return nil, exception.BadRequest("status must be DRAFT, PUBLISHED, COMPLETED, or CANCELLED")
	}

	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, exception.NotFound("event not found")
	}
	if ev.OrganizerID != organizerID {
		return nil, exception.Forbidden("only the lead organizer can change event status")
	}

	if status == entity.EventStatusPublished && u.organizerRepo != nil {
		profile, err := u.organizerRepo.FindByUserID(ctx, organizerID)
		if err != nil {
			return nil, err
		}
		if profile == nil {
			return nil, exception.UnprocessableEntity("organizer profile must be completed before publishing an event")
		}
	}

	if err := u.eventRepo.UpdateStatus(ctx, eventID, status); err != nil {
		return nil, err
	}
	ev.Status = status
	return u.buildEventDetail(ctx, ev)
}

// GetArtists returns paginated master artists (F-4.3).
func (u *EventUsecase) GetArtists(ctx context.Context, req request.GetArtistListRequest) (*response.PageResponse[response.EventArtistResponse], error) {
	req.Normalize(20)

	filter := repository.ArtistFilter{
		Search:    req.Search,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Limit:     req.PageSize,
		Offset:    model.CalculateOffset(req.Page, req.PageSize),
	}

	artists, total, err := u.eventRepo.FindAllArtists(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]response.EventArtistResponse, 0, len(artists))
	for _, a := range artists {
		items = append(items, response.EventArtistResponse{
			ID:        a.ID,
			Name:      a.Name,
			Bio:       a.Bio,
			AvatarURL: a.AvatarURL,
		})
	}
	return &response.PageResponse[response.EventArtistResponse]{
		Items:      items,
		Pagination: model.BuildPagination(total, req.Page, req.PageSize),
	}, nil
}

// CreateArtist creates a master artist profile with optional avatar upload (F-4.3).
func (u *EventUsecase) CreateArtist(
	ctx context.Context,
	req request.CreateArtistRequest,
	avatarHeader *multipart.FileHeader,
) (*response.EventArtistResponse, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 || len(name) > 150 {
		return nil, exception.BadRequest("name must be between 2 and 150 characters")
	}

	var bioPtr *string
	if bio := strings.TrimSpace(req.Bio); bio != "" {
		bioPtr = &bio
	}

	avatarURL, err := u.uploadOptionalImage(ctx, avatarHeader, "artists")
	if err != nil {
		return nil, err
	}

	artist := &entity.Artist{
		ID:        uuid.New(),
		Name:      name,
		Bio:       bioPtr,
		AvatarURL: avatarURL,
	}
	if err := u.eventRepo.CreateArtist(ctx, artist); err != nil {
		return nil, err
	}

	return &response.EventArtistResponse{
		ID:        artist.ID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}, nil
}

func (u *EventUsecase) canManageEventLineup(ctx context.Context, actorID uuid.UUID, actorRoleID *int, ev *entity.Event) (bool, error) {
	if actorRoleID != nil && u.roleRepo != nil {
		role, err := u.roleRepo.FindByID(ctx, *actorRoleID)
		if err != nil {
			return false, err
		}
		if role != nil {
			roleName := strings.ToUpper(role.Name)
			if roleName == "SUPER_ADMIN" || roleName == "ADMIN" {
				return true, nil
			}
		}
	}
	return ev.OrganizerID == actorID, nil
}

// AddEventArtist attaches an existing artist to an event's lineup (F-4.3).
func (u *EventUsecase) AddEventArtist(
	ctx context.Context,
	actorID uuid.UUID,
	actorRoleID *int,
	eventID uuid.UUID,
	req request.AddEventArtistRequest,
) (*response.EventArtistResponse, error) {
	artistID, err := uuid.Parse(strings.TrimSpace(req.ArtistID))
	if err != nil {
		return nil, exception.BadRequest("invalid artistId format")
	}

	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, exception.NotFound("event not found")
	}

	allowed, err := u.canManageEventLineup(ctx, actorID, actorRoleID, ev)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, exception.Forbidden("you do not have permission to manage this event lineup")
	}

	artist, err := u.eventRepo.FindArtistByID(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if artist == nil {
		return nil, exception.NotFound("artist not found")
	}

	exists, err := u.eventRepo.IsEventArtist(ctx, eventID, artistID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exception.Conflict("artist is already in the event lineup")
	}

	if err := u.eventRepo.AddEventArtist(ctx, eventID, artistID); err != nil {
		return nil, err
	}

	return &response.EventArtistResponse{
		ID:        artist.ID,
		Name:      artist.Name,
		Bio:       artist.Bio,
		AvatarURL: artist.AvatarURL,
	}, nil
}

// RemoveEventArtist removes an artist from an event's lineup (F-4.3).
func (u *EventUsecase) RemoveEventArtist(
	ctx context.Context,
	actorID uuid.UUID,
	actorRoleID *int,
	eventID uuid.UUID,
	artistID uuid.UUID,
) error {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if ev == nil {
		return exception.NotFound("event not found")
	}

	allowed, err := u.canManageEventLineup(ctx, actorID, actorRoleID, ev)
	if err != nil {
		return err
	}
	if !allowed {
		return exception.Forbidden("you do not have permission to manage this event lineup")
	}

	removed, err := u.eventRepo.RemoveEventArtist(ctx, eventID, artistID)
	if err != nil {
		return err
	}
	if !removed {
		return exception.NotFound("artist not found in event lineup")
	}
	return nil
}

// GetEventTeam lists assigned CO_ORGANIZER and STAFF members for an event (F-4.4).
func (u *EventUsecase) GetEventTeam(ctx context.Context, organizerID, eventID uuid.UUID) ([]response.EventTeamMemberResponse, error) {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, exception.NotFound("event not found")
	}
	if ev.OrganizerID != organizerID {
		return nil, exception.Forbidden("only the lead organizer can view event team")
	}

	members, err := u.eventRepo.FindTeamByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	result := make([]response.EventTeamMemberResponse, 0, len(members))
	for _, m := range members {
		item := response.EventTeamMemberResponse{
			ID:          m.ID,
			EventID:     m.EventID,
			UserID:      m.UserID,
			Name:        m.Name,
			Email:       m.Email,
			OrgName:     m.OrgName,
			AvatarURL:   m.AvatarURL,
			RoleInEvent: m.RoleInEvent,
			Status:      m.Status,
			AssignedAt:  m.AssignedAt,
		}
		if item.Name == "" && u.userRepo != nil {
			if user, err := u.userRepo.FindByID(ctx, m.UserID); err == nil && user != nil {
				item.Name = user.Name
				item.Email = user.Email
				item.AvatarURL = user.AvatarURL
				item.Status = user.Status
			}
		}
		if item.OrgName == nil && item.RoleInEvent == entity.EventRoleCoOrganizer && u.organizerRepo != nil {
			if profile, err := u.organizerRepo.FindByUserID(ctx, m.UserID); err == nil && profile != nil {
				orgName := profile.OrgName
				item.OrgName = &orgName
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// AssignEventTeam assigns an active STAFF (owned by lead organizer) or CO_ORGANIZER to an event (F-4.4).
func (u *EventUsecase) AssignEventTeam(
	ctx context.Context,
	organizerID uuid.UUID,
	eventID uuid.UUID,
	req request.AssignEventTeamRequest,
) (*response.EventTeamMemberResponse, error) {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, exception.NotFound("event not found")
	}
	if ev.OrganizerID != organizerID {
		return nil, exception.Forbidden("only the lead organizer can assign event team members")
	}

	targetUserID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		return nil, exception.BadRequest("invalid userId format")
	}

	roleInEvent := strings.ToUpper(strings.TrimSpace(req.RoleInEvent))
	if roleInEvent != entity.EventRoleCoOrganizer && roleInEvent != entity.EventRoleStaff {
		return nil, exception.BadRequest("roleInEvent must be CO_ORGANIZER or STAFF")
	}

	if u.userRepo == nil {
		return nil, exception.Internal("user repository not configured", nil)
	}
	targetUser, err := u.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, exception.NotFound("target user not found")
	}
	if targetUser.Status != entity.UserStatusActive {
		return nil, exception.UnprocessableEntity("target user must have ACTIVE status")
	}

	targetRoleName := ""
	if targetUser.RoleID != nil && u.roleRepo != nil {
		role, err := u.roleRepo.FindByID(ctx, *targetUser.RoleID)
		if err != nil {
			return nil, err
		}
		if role != nil {
			targetRoleName = strings.ToUpper(role.Name)
		}
	}

	var orgNamePtr *string
	switch roleInEvent {
	case entity.EventRoleStaff:
		if targetRoleName != "STAFF" {
			return nil, exception.UnprocessableEntity("target user must have STAFF role")
		}
		if u.organizerRepo == nil {
			return nil, exception.Internal("organizer repository not configured", nil)
		}
		owned, err := u.organizerRepo.IsStaffOfOrganizer(ctx, organizerID, targetUserID)
		if err != nil {
			return nil, err
		}
		if !owned {
			return nil, exception.Forbidden("target staff does not belong to your organization")
		}
	case entity.EventRoleCoOrganizer:
		if targetUserID == organizerID {
			return nil, exception.BadRequest("lead organizer cannot be assigned as CO_ORGANIZER")
		}
		if targetRoleName != "ORGANIZER" {
			return nil, exception.UnprocessableEntity("target user must have ORGANIZER role")
		}
		if u.organizerRepo != nil {
			if profile, err := u.organizerRepo.FindByUserID(ctx, targetUserID); err == nil && profile != nil {
				orgName := profile.OrgName
				orgNamePtr = &orgName
			}
		}
	}

	exists, err := u.eventRepo.IsTeamMember(ctx, eventID, targetUserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, exception.Conflict("user is already assigned to this event team")
	}

	member := &entity.EventOrganizer{
		ID:          uuid.New(),
		EventID:     eventID,
		UserID:      targetUserID,
		RoleInEvent: roleInEvent,
		AssignedAt:  time.Now(),
	}
	if err := u.eventRepo.AddTeamMember(ctx, member); err != nil {
		return nil, err
	}

	return &response.EventTeamMemberResponse{
		ID:          member.ID,
		EventID:     member.EventID,
		UserID:      member.UserID,
		Name:        targetUser.Name,
		Email:       targetUser.Email,
		OrgName:     orgNamePtr,
		AvatarURL:   targetUser.AvatarURL,
		RoleInEvent: member.RoleInEvent,
		Status:      targetUser.Status,
		AssignedAt:  member.AssignedAt,
	}, nil
}

// RemoveEventTeam unassigns a CO_ORGANIZER or STAFF member from an event (F-4.4).
func (u *EventUsecase) RemoveEventTeam(ctx context.Context, organizerID, eventID, targetUserID uuid.UUID) error {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if ev == nil {
		return exception.NotFound("event not found")
	}
	if ev.OrganizerID != organizerID {
		return exception.Forbidden("only the lead organizer can remove event team members")
	}

	removed, err := u.eventRepo.RemoveTeamMember(ctx, eventID, targetUserID)
	if err != nil {
		return err
	}
	if !removed {
		return exception.NotFound("team member not found in this event")
	}
	return nil
}

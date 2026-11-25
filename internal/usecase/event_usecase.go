package usecase

import (
	"context"
	"strings"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/domain/payload/response"
	"concert-go/internal/exception"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

type EventUsecase struct {
	eventRepo     repository.EventRepository
	organizerRepo repository.OrganizerRepository
	userRepo      repository.UserRepository
}

func NewEventUsecase(
	eventRepo repository.EventRepository,
	organizerRepo repository.OrganizerRepository,
	userRepo repository.UserRepository,
) *EventUsecase {
	return &EventUsecase{
		eventRepo:     eventRepo,
		organizerRepo: organizerRepo,
		userRepo:      userRepo,
	}
}

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
		Pagination: model.BuildPagination(total, req.Page, req.PageSize),
	}, nil
}

func (u *EventUsecase) GetPublicEventDetail(ctx context.Context, eventID uuid.UUID) (*response.EventDetailResponse, error) {
	ev, err := u.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if ev == nil || ev.Status != entity.EventStatusPublished {
		return nil, exception.NotFound("event not found")
	}

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
		coOrgs = append(coOrgs, response.EventCoOrganizerResponse{
			UserID:     co.UserID,
			Name:       co.Name,
			OrgName:    co.OrgName,
			LogoURL:    co.LogoURL,
			AssignedAt: co.AssignedAt,
		})
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

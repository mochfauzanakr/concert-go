package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/domain/model"
	"concert-go/internal/domain/payload/request"
	"concert-go/internal/exception"
	"concert-go/internal/repository"
	"concert-go/internal/repository/mock"
	"concert-go/internal/usecase"

	"github.com/google/uuid"
)

func setupEventUsecase() (*usecase.EventUsecase, *mock.EventRepositoryMock, repository.OrganizerRepository, repository.UserRepository) {
	eventRepo := mock.NewEventRepositoryMock()
	organizerRepo := mock.NewOrganizerRepositoryMock()
	userRepo := mock.NewUserRepositoryMock()
	uc := usecase.NewEventUsecase(eventRepo, organizerRepo, userRepo)
	return uc, eventRepo, organizerRepo, userRepo
}

func TestEventUsecase_GetPublicEvents(t *testing.T) {
	ctx := context.Background()
	uc, eventRepo, _, _ := setupEventUsecase()

	orgID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	pubRockJkt := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   orgID,
		Name:          "Jakarta Rock Fest",
		Description:   "Annual rock festival",
		Category:      "Rock",
		City:          "Jakarta",
		Location:      "GBK Senayan",
		StartTime:     now.Add(10 * 24 * time.Hour),
		EndTime:       now.Add(10*24*time.Hour + 5*time.Hour),
		Status:        entity.EventStatusPublished,
		InventoryMode: entity.InventoryModeOnlineOnly,
	}
	pubJazzBdg := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   orgID,
		Name:          "Bandung Jazz Night",
		Description:   "Smooth jazz under the stars",
		Category:      "Jazz",
		City:          "Bandung",
		Location:      "Dago Tea House",
		StartTime:     now.Add(5 * 24 * time.Hour),
		EndTime:       now.Add(5*24*time.Hour + 4*time.Hour),
		Status:        entity.EventStatusPublished,
		InventoryMode: entity.InventoryModeDualChannel,
	}
	pubPopSby := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   orgID,
		Name:          "Surabaya Pop Carnival",
		Description:   "Pop music carnival",
		Category:      "Pop",
		City:          "Surabaya",
		Location:      "Jatim Expo",
		StartTime:     now.Add(20 * 24 * time.Hour),
		EndTime:       now.Add(20*24*time.Hour + 6*time.Hour),
		Status:        entity.EventStatusPublished,
		InventoryMode: entity.InventoryModeOnlineOnly,
	}
	draftEvent := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   orgID,
		Name:          "Secret Draft Gig",
		Description:   "Not published yet",
		Category:      "Rock",
		City:          "Jakarta",
		Location:      "Secret Venue",
		StartTime:     now.Add(7 * 24 * time.Hour),
		EndTime:       now.Add(7*24*time.Hour + 3*time.Hour),
		Status:        entity.EventStatusDraft,
		InventoryMode: entity.InventoryModeOnlineOnly,
	}
	cancelledEvent := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   orgID,
		Name:          "Cancelled Indie Show",
		Description:   "Cancelled",
		Category:      "Indie",
		City:          "Jakarta",
		Location:      "Senayan",
		StartTime:     now.Add(8 * 24 * time.Hour),
		EndTime:       now.Add(8*24*time.Hour + 3*time.Hour),
		Status:        entity.EventStatusCancelled,
		InventoryMode: entity.InventoryModeOnlineOnly,
	}

	eventRepo.SeedEvent(pubRockJkt)
	eventRepo.SeedEvent(pubJazzBdg)
	eventRepo.SeedEvent(pubPopSby)
	eventRepo.SeedEvent(draftEvent)
	eventRepo.SeedEvent(cancelledEvent)

	artist := &entity.Artist{
		ID:   uuid.New(),
		Name: "Sheila On 7",
	}
	eventRepo.SeedArtist(artist)
	eventRepo.SeedEventArtist(pubJazzBdg.ID, artist.ID)

	t.Run("returns only PUBLISHED events", func(t *testing.T) {
		res, err := uc.GetPublicEvents(ctx, request.GetPublicEventListRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Pagination.TotalRecords != 3 {
			t.Fatalf("expected 3 published events, got %d", res.Pagination.TotalRecords)
		}
		for _, item := range res.Items {
			if item.Status != entity.EventStatusPublished {
				t.Errorf("expected only PUBLISHED status, got %s", item.Status)
			}
		}
	})

	t.Run("filters by city and category case-insensitively", func(t *testing.T) {
		res, err := uc.GetPublicEvents(ctx, request.GetPublicEventListRequest{
			City:     "jakarta",
			Category: "ROCK",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 1 || res.Items[0].ID != pubRockJkt.ID {
			t.Fatalf("expected only Jakarta Rock Fest, got %+v", res.Items)
		}
	})

	t.Run("search matches artist name in lineup", func(t *testing.T) {
		res, err := uc.GetPublicEvents(ctx, request.GetPublicEventListRequest{
			BaseListParams: model.BaseListParams{Search: "sheila"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 1 || res.Items[0].ID != pubJazzBdg.ID {
			t.Fatalf("expected Bandung Jazz Night via artist search, got %+v", res.Items)
		}
	})

	t.Run("filters by startDate and endDate on startTime", func(t *testing.T) {
		from := now.Add(9 * 24 * time.Hour)
		to := now.Add(15 * 24 * time.Hour)
		res, err := uc.GetPublicEvents(ctx, request.GetPublicEventListRequest{
			BaseListParams: model.BaseListParams{
				StartDate: &from,
				EndDate:   &to,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 1 || res.Items[0].ID != pubRockJkt.ID {
			t.Fatalf("expected only Jakarta Rock Fest in date range, got %+v", res.Items)
		}
	})

	t.Run("supports sorting by startTime ASC and pagination", func(t *testing.T) {
		res, err := uc.GetPublicEvents(ctx, request.GetPublicEventListRequest{
			BaseListParams: model.BaseListParams{
				Page:      1,
				PageSize:  2,
				SortBy:    "startTime",
				SortOrder: "ASC",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 2 {
			t.Fatalf("expected 2 items on page 1, got %d", len(res.Items))
		}
		if res.Items[0].ID != pubJazzBdg.ID || res.Items[1].ID != pubRockJkt.ID {
			t.Errorf("unexpected sort order: got %s then %s", res.Items[0].Name, res.Items[1].Name)
		}
		if res.Pagination.TotalPage != 2 || res.Pagination.NextPage == nil || *res.Pagination.NextPage != 2 {
			t.Errorf("unexpected pagination metadata: %+v", res.Pagination)
		}
	})
}

func TestEventUsecase_GetPublicEventDetail(t *testing.T) {
	ctx := context.Background()
	uc, eventRepo, organizerRepo, userRepo := setupEventUsecase()

	leadUserID := uuid.New()
	logo := "https://cdn.example.com/org-logo.png"
	_ = organizerRepo.Upsert(ctx, &entity.OrganizerProfile{
		UserID:  leadUserID,
		OrgName: "Nusantara Live",
		LogoURL: &logo,
	})

	banner := "https://cdn.example.com/banner.jpg"
	publishedEvent := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   leadUserID,
		Name:          "Nusantara Soundscape",
		Description:   "Full lineup festival",
		BannerURL:     &banner,
		Category:      "Festival",
		City:          "Jakarta",
		Location:      "JIExpo Kemayoran",
		StartTime:     time.Now().Add(14 * 24 * time.Hour),
		EndTime:       time.Now().Add(14*24*time.Hour + 8*time.Hour),
		Status:        entity.EventStatusPublished,
		InventoryMode: entity.InventoryModeDualChannel,
	}
	draftEvent := &entity.Event{
		ID:            uuid.New(),
		OrganizerID:   leadUserID,
		Name:          "Private Draft Event",
		Description:   "Draft",
		Category:      "Indie",
		City:          "Bandung",
		Location:      "Studio",
		StartTime:     time.Now().Add(20 * 24 * time.Hour),
		EndTime:       time.Now().Add(20*24*time.Hour + 3*time.Hour),
		Status:        entity.EventStatusDraft,
		InventoryMode: entity.InventoryModeOnlineOnly,
	}

	eventRepo.SeedEvent(publishedEvent)
	eventRepo.SeedEvent(draftEvent)

	bio := "Legendary indie band"
	artist := &entity.Artist{
		ID:   uuid.New(),
		Name: "Efek Rumah Kaca",
		Bio:  &bio,
	}
	eventRepo.SeedArtist(artist)
	eventRepo.SeedEventArtist(publishedEvent.ID, artist.ID)

	coOrgID := uuid.New()
	eventRepo.SeedCoOrganizer(publishedEvent.ID, &repository.EventCoOrganizer{
		UserID:     coOrgID,
		Name:       "CoHost Pic",
		OrgName:    "Partner Promotor",
		AssignedAt: time.Now(),
	})

	t.Run("returns 404 for unknown event ID", func(t *testing.T) {
		_, err := uc.GetPublicEventDetail(ctx, uuid.New())
		var appErr *exception.AppException
		if !errors.As(err, &appErr) || appErr.Code != 404 {
			t.Fatalf("expected 404 AppError, got %v", err)
		}
	})

	t.Run("returns 404 for non-published event", func(t *testing.T) {
		_, err := uc.GetPublicEventDetail(ctx, draftEvent.ID)
		var appErr *exception.AppException
		if !errors.As(err, &appErr) || appErr.Code != 404 {
			t.Fatalf("expected 404 AppError for DRAFT event, got %v", err)
		}
	})

	t.Run("returns full detail with lead organizer, co-organizers, artists, and ticket summary", func(t *testing.T) {
		detail, err := uc.GetPublicEventDetail(ctx, publishedEvent.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.ID != publishedEvent.ID || detail.Name != publishedEvent.Name {
			t.Errorf("unexpected event identity: %+v", detail)
		}
		if detail.LeadOrganizer == nil || detail.LeadOrganizer.OrgName != "Nusantara Live" {
			t.Fatalf("expected LeadOrganizer Nusantara Live, got %+v", detail.LeadOrganizer)
		}
		if len(detail.CoOrganizers) != 1 || detail.CoOrganizers[0].OrgName != "Partner Promotor" {
			t.Fatalf("expected 1 CoOrganizer Partner Promotor, got %+v", detail.CoOrganizers)
		}
		if len(detail.Artists) != 1 || detail.Artists[0].Name != "Efek Rumah Kaca" {
			t.Fatalf("expected 1 Artist Efek Rumah Kaca, got %+v", detail.Artists)
		}
		if detail.TicketSummary.TotalAvailable != 0 || detail.TicketSummary.LowestPrice != nil || detail.TicketSummary.SoldOut {
			t.Errorf("unexpected placeholder TicketSummary: %+v", detail.TicketSummary)
		}
	})

	t.Run("falls back to user name when organizer profile is missing", func(t *testing.T) {
		fallbackOrgID := uuid.New()
		_ = userRepo.Create(ctx, &entity.User{
			ID:     fallbackOrgID,
			Name:   "Fallback Organizer",
			Email:  "fallback@example.com",
			Status: entity.UserStatusActive,
		})
		ev := &entity.Event{
			ID:            uuid.New(),
			OrganizerID:   fallbackOrgID,
			Name:          "Fallback Show",
			Description:   "Show",
			Category:      "Pop",
			City:          "Bali",
			Location:      "GWK",
			StartTime:     time.Now().Add(30 * 24 * time.Hour),
			EndTime:       time.Now().Add(30*24*time.Hour + 4*time.Hour),
			Status:        entity.EventStatusPublished,
			InventoryMode: entity.InventoryModeOnlineOnly,
		}
		eventRepo.SeedEvent(ev)

		detail, err := uc.GetPublicEventDetail(ctx, ev.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.LeadOrganizer == nil || detail.LeadOrganizer.OrgName != "Fallback Organizer" {
			t.Fatalf("expected fallback LeadOrganizer name, got %+v", detail.LeadOrganizer)
		}
	})
}

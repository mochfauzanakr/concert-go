package mock

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"concert-go/internal/domain/entity"
	"concert-go/internal/repository"

	"github.com/google/uuid"
)

// EventRepositoryMock is an in-memory thread-safe implementation of repository.EventRepository.
type EventRepositoryMock struct {
	mu           sync.RWMutex
	events       map[uuid.UUID]*entity.Event
	artists      map[uuid.UUID]*entity.Artist
	eventArtists map[uuid.UUID][]uuid.UUID
	coOrganizers map[uuid.UUID][]*repository.EventCoOrganizer
}

// NewEventRepositoryMock creates a new in-memory mock EventRepository.
func NewEventRepositoryMock() *EventRepositoryMock {
	return &EventRepositoryMock{
		events:       make(map[uuid.UUID]*entity.Event),
		artists:      make(map[uuid.UUID]*entity.Artist),
		eventArtists: make(map[uuid.UUID][]uuid.UUID),
		coOrganizers: make(map[uuid.UUID][]*repository.EventCoOrganizer),
	}
}

// SeedEvent inserts or replaces an event in the mock store for testing.
func (m *EventRepositoryMock) SeedEvent(ev *entity.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	now := time.Now()
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = now
	}
	if ev.UpdatedAt.IsZero() {
		ev.UpdatedAt = now
	}
	copyEv := *ev
	m.events[ev.ID] = &copyEv
}

// SeedArtist inserts or replaces an artist in the mock store for testing.
func (m *EventRepositoryMock) SeedArtist(artist *entity.Artist) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if artist.ID == uuid.Nil {
		artist.ID = uuid.New()
	}
	now := time.Now()
	if artist.CreatedAt.IsZero() {
		artist.CreatedAt = now
	}
	if artist.UpdatedAt.IsZero() {
		artist.UpdatedAt = now
	}
	copyArtist := *artist
	m.artists[artist.ID] = &copyArtist
}

// SeedEventArtist links an artist to an event in the mock store for testing.
func (m *EventRepositoryMock) SeedEventArtist(eventID, artistID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.eventArtists[eventID] {
		if existing == artistID {
			return
		}
	}
	m.eventArtists[eventID] = append(m.eventArtists[eventID], artistID)
}

// SeedCoOrganizer attaches a co-organizer record to an event in the mock store for testing.
func (m *EventRepositoryMock) SeedCoOrganizer(eventID uuid.UUID, coOrg *repository.EventCoOrganizer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	copyCo := *coOrg
	m.coOrganizers[eventID] = append(m.coOrganizers[eventID], &copyCo)
}

func (m *EventRepositoryMock) FindAll(ctx context.Context, filter repository.EventFilter) ([]*entity.Event, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))
	cityLower := strings.ToLower(strings.TrimSpace(filter.City))
	categoryLower := strings.ToLower(strings.TrimSpace(filter.Category))

	var filtered []*entity.Event
	for _, ev := range m.events {
		if ev.DeletedAt.Valid {
			continue
		}
		if filter.Status != "" && ev.Status != filter.Status {
			continue
		}
		if filter.ManagedByUserID != nil && !m.isEventManagerLocked(ev, *filter.ManagedByUserID) {
			continue
		}
		if cityLower != "" && strings.ToLower(ev.City) != cityLower {
			continue
		}
		if categoryLower != "" && strings.ToLower(ev.Category) != categoryLower {
			continue
		}
		if filter.StartDate != nil && ev.StartTime.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && ev.StartTime.After(*filter.EndDate) {
			continue
		}
		if searchLower != "" {
			matched := strings.Contains(strings.ToLower(ev.Name), searchLower)
			if !matched {
				for _, artistID := range m.eventArtists[ev.ID] {
					if artist, ok := m.artists[artistID]; ok && strings.Contains(strings.ToLower(artist.Name), searchLower) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		copyEv := *ev
		filtered = append(filtered, &copyEv)
	}

	asc := strings.EqualFold(strings.TrimSpace(filter.SortOrder), "ASC")
	sortBy := strings.ToLower(strings.TrimSpace(filter.SortBy))
	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "starttime", "start_time":
			less = filtered[i].StartTime.Before(filtered[j].StartTime)
		case "name":
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		case "city":
			less = strings.ToLower(filtered[i].City) < strings.ToLower(filtered[j].City)
		case "category":
			less = strings.ToLower(filtered[i].Category) < strings.ToLower(filtered[j].Category)
		default:
			less = filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		if asc {
			return less
		}
		return !less
	})

	total := len(filtered)
	start := filter.Offset
	if start > total {
		start = total
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (m *EventRepositoryMock) isEventManagerLocked(ev *entity.Event, userID uuid.UUID) bool {
	if ev.OrganizerID == userID {
		return true
	}
	for _, co := range m.coOrganizers[ev.ID] {
		if co.UserID == userID {
			return true
		}
	}
	return false
}

func (m *EventRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ev, exists := m.events[id]; exists && !ev.DeletedAt.Valid {
		copyEv := *ev
		return &copyEv, nil
	}
	return nil, nil
}

func (m *EventRepositoryMock) Create(ctx context.Context, event *entity.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Status == "" {
		event.Status = entity.EventStatusDraft
	}
	if event.InventoryMode == "" {
		event.InventoryMode = entity.InventoryModeOnlineOnly
	}
	event.CreatedAt = now
	event.UpdatedAt = now
	copyEv := *event
	m.events[event.ID] = &copyEv
	return nil
}

func (m *EventRepositoryMock) Update(ctx context.Context, event *entity.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.events[event.ID]
	if !ok || existing.DeletedAt.Valid {
		return nil
	}
	event.CreatedAt = existing.CreatedAt
	event.UpdatedAt = time.Now()
	copyEv := *event
	m.events[event.ID] = &copyEv
	return nil
}

func (m *EventRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ev, ok := m.events[id]; ok && !ev.DeletedAt.Valid {
		ev.Status = status
		ev.UpdatedAt = time.Now()
	}
	return nil
}

func (m *EventRepositoryMock) FindArtistsByEventID(ctx context.Context, eventID uuid.UUID) ([]*entity.Artist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	artistIDs := m.eventArtists[eventID]
	artists := make([]*entity.Artist, 0, len(artistIDs))
	for _, artistID := range artistIDs {
		if artist, ok := m.artists[artistID]; ok {
			copyArtist := *artist
			artists = append(artists, &copyArtist)
		}
	}
	sort.Slice(artists, func(i, j int) bool {
		return strings.ToLower(artists[i].Name) < strings.ToLower(artists[j].Name)
	})
	return artists, nil
}

func (m *EventRepositoryMock) FindAllArtists(ctx context.Context, filter repository.ArtistFilter) ([]*entity.Artist, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))
	var filtered []*entity.Artist
	for _, a := range m.artists {
		if searchLower != "" && !strings.Contains(strings.ToLower(a.Name), searchLower) {
			continue
		}
		if filter.StartDate != nil && a.CreatedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && a.CreatedAt.After(*filter.EndDate) {
			continue
		}
		copyArtist := *a
		filtered = append(filtered, &copyArtist)
	}

	asc := strings.EqualFold(strings.TrimSpace(filter.SortOrder), "ASC")
	sortBy := strings.ToLower(strings.TrimSpace(filter.SortBy))
	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		default:
			less = filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		if asc {
			return less
		}
		return !less
	})

	total := len(filtered)
	start := filter.Offset
	if start > total {
		start = total
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	end := start + limit
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (m *EventRepositoryMock) FindArtistByID(ctx context.Context, id uuid.UUID) (*entity.Artist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if a, ok := m.artists[id]; ok {
		copyArtist := *a
		return &copyArtist, nil
	}
	return nil, nil
}

func (m *EventRepositoryMock) CreateArtist(ctx context.Context, artist *entity.Artist) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if artist.ID == uuid.Nil {
		artist.ID = uuid.New()
	}
	artist.CreatedAt = now
	artist.UpdatedAt = now
	copyArtist := *artist
	m.artists[artist.ID] = &copyArtist
	return nil
}

func (m *EventRepositoryMock) IsEventArtist(ctx context.Context, eventID, artistID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, id := range m.eventArtists[eventID] {
		if id == artistID {
			return true, nil
		}
	}
	return false, nil
}

func (m *EventRepositoryMock) AddEventArtist(ctx context.Context, eventID, artistID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range m.eventArtists[eventID] {
		if id == artistID {
			return nil
		}
	}
	m.eventArtists[eventID] = append(m.eventArtists[eventID], artistID)
	return nil
}

func (m *EventRepositoryMock) RemoveEventArtist(ctx context.Context, eventID, artistID uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.eventArtists[eventID]
	for i, id := range list {
		if id == artistID {
			m.eventArtists[eventID] = append(list[:i], list[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (m *EventRepositoryMock) FindCoOrganizersByEventID(ctx context.Context, eventID uuid.UUID) ([]*repository.EventCoOrganizer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.coOrganizers[eventID]
	result := make([]*repository.EventCoOrganizer, 0, len(list))
	for _, item := range list {
		copyItem := *item
		result = append(result, &copyItem)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].AssignedAt.Before(result[j].AssignedAt)
	})
	return result, nil
}

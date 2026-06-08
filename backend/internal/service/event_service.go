package service

import (
	"context"
	"time"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

type CreateEventParams struct {
	Title       string
	Description string
	Place       string
	ShowTime    time.Time
	ImgURL      string
	OrganizerID string
	CategoryID  *string
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) CreateEvent(ctx context.Context, params CreateEventParams) (*repository.Event, error) {
	return s.repo.CreateEvent(ctx, repository.CreateEventParams{
		Title:       params.Title,
		Description: params.Description,
		Place:       params.Place,
		ShowTime:    params.ShowTime,
		ImgURL:      params.ImgURL,
		OrganizerID: params.OrganizerID,
		CategoryID:  params.CategoryID,
	})
}

func (s *EventService) GetEvents(ctx context.Context) ([]repository.Event, error) {
	return s.repo.GetEvents(ctx)
}

func (s *EventService) GetEventByID(ctx context.Context, id string) (*repository.Event, error) {
	return s.repo.GetEventByID(ctx, id)
}

package service

import (
	"context"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
)

type TicketPoolService struct {
	repo *repository.TicketPoolRepository
}

type CreateTicketPoolParams struct {
	ZoneName      string
	Price         float64
	TotalCapacity int
	DisplayName   string
	EventID       string
}

func NewTicketPoolService(repo *repository.TicketPoolRepository) *TicketPoolService {
	return &TicketPoolService{repo: repo}
}

func (s *TicketPoolService) CreateTicketPool(ctx context.Context, params CreateTicketPoolParams) (*repository.TicketPool, error) {
	return s.repo.CreateTicketPool(ctx, repository.CreateTicketPoolParams{
		ZoneName:      params.ZoneName,
		Price:         params.Price,
		TotalCapacity: params.TotalCapacity,
		DisplayName:   params.DisplayName,
		EventID:       params.EventID,
	})
}

func (s *TicketPoolService) GetTicketPoolsByEventID(ctx context.Context, eventID string) ([]repository.TicketPool, error) {
	return s.repo.GetTicketPoolsByEventID(ctx, eventID)
}

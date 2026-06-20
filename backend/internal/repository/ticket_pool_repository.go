package repository

import (
	"context"
	"database/sql"
	"time"
)

type TicketPoolRepository struct {
	db *sql.DB
}

func NewTicketPoolRepository(db *sql.DB) *TicketPoolRepository {
	return &TicketPoolRepository{db: db}
}

type CreateTicketPoolParams struct {
	ZoneName      string
	Price         float64
	TotalCapacity int
	DisplayName   string
	EventID       string
}

type TicketPool struct {
	ID                string    `json:"id"`
	ZoneName          string    `json:"zone_name"`
	Price             float64   `json:"price"`
	TotalCapacity     int       `json:"total_capacity"`
	AvailableCapacity int       `json:"available_capacity"`
	DisplayName       string    `json:"display_name"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Version           int       `json:"version"`
	EventID           string    `json:"event_id"`
}

func (r *TicketPoolRepository) CreateTicketPool(ctx context.Context, params CreateTicketPoolParams) (*TicketPool, error){
	var ticketPool TicketPool
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO ticket_pools(zone_name, price, total_capacity, available_capacity, display_name, event_id)
		VALUES ($1, $2, $3, $3, $4, $5)
		RETURNING id, zone_name, price, total_capacity, available_capacity, display_name, created_at, updated_at, version, event_id
	`, params.ZoneName, params.Price, params.TotalCapacity, params.DisplayName, params.EventID,
	).Scan(&ticketPool.ID, &ticketPool.ZoneName, &ticketPool.Price, &ticketPool.TotalCapacity, 
		&ticketPool.AvailableCapacity, &ticketPool.DisplayName, &ticketPool.CreatedAt, &ticketPool.UpdatedAt, 
		&ticketPool.Version, &ticketPool.EventID, 
	)
	return &ticketPool, err
}

func (r *TicketPoolRepository) GetTicketPoolsByEventID (ctx context.Context, eventID string) ([]TicketPool, error){
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, zone_name, price, total_capacity, available_capacity, display_name, created_at, updated_at, version, event_id
		FROM ticket_pools WHERE event_id = $1;
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ticketPools []TicketPool
	for rows.Next() {
		var ticketPool TicketPool
		if err := rows.Scan(&ticketPool.ID, &ticketPool.ZoneName, &ticketPool.Price, &ticketPool.TotalCapacity, 
			&ticketPool.AvailableCapacity, &ticketPool.DisplayName, &ticketPool.CreatedAt, &ticketPool.UpdatedAt, 
			&ticketPool.Version, &ticketPool.EventID, 
		); err != nil {
			return nil, err
		}
		ticketPools = append(ticketPools, ticketPool)
	}
	return ticketPools, err
}
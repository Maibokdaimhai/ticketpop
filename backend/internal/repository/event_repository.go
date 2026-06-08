package repository

import (
	"context"
	"database/sql"
	"time"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
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

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Place       string    `json:"place"`
	ShowTime    time.Time `json:"show_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ImgURL      string    `json:"img_url"`
	OrganizerID string    `json:"organizer_id"`
	CategoryID  *string   `json:"category_id"`
}

func (r *EventRepository) CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error) {
	var event Event
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO events(title, description, place, show_time, img_url, organizer_id, category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, place, show_time, img_url, organizer_id, category_id, created_at, updated_at;
	`, params.Title, params.Description, params.Place, params.ShowTime,
		params.ImgURL, params.OrganizerID, params.CategoryID,
	).Scan(&event.ID, &event.Title, &event.Description, &event.Place,
		&event.ShowTime, &event.ImgURL, &event.OrganizerID, &event.CategoryID,
		&event.CreatedAt, &event.UpdatedAt,
	)
	return &event, err
}

func (r *EventRepository) GetEvents(ctx context.Context) ([]Event, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, place, show_time, created_at, updated_at, img_url, organizer_id, category_id
		FROM events 
		WHERE deleted_at IS null;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.Place,
			&event.ShowTime, &event.CreatedAt, &event.UpdatedAt, &event.ImgURL,
			&event.OrganizerID, &event.CategoryID,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, err
}

func (r *EventRepository) GetEventByID(ctx context.Context, id string) (*Event, error) {
	var event Event
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, place, show_time, created_at, updated_at, img_url, organizer_id, category_id
		FROM events
		WHERE deleted_at IS null AND id = $1;
	`, id).Scan(&event.ID, &event.Title, &event.Description, &event.Place,
		&event.ShowTime, &event.CreatedAt, &event.UpdatedAt, &event.ImgURL,
		&event.OrganizerID, &event.CategoryID,
	)
	return &event, err
}

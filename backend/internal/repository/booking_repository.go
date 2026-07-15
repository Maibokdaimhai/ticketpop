package repository

import (
	"context"
	"database/sql"
	"time"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

type CreateBookingParams struct {
	UserID     string
	EventID    string
	TotalPrice float64
	ExpiresAt  time.Time
}

type Booking struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UserID     string    `json:"user_id"`
	EventID    string    `json:"event_id"`
}

type CreateTicketParams struct {
	SeatNumber   *string
	TicketPoolID string
	BookingID    string
}

type Ticket struct {
	ID           string    `json:"id"`
	SeatNumber   *string   `json:"seat_number"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TicketPoolID string    `json:"ticket_pool_id"`
	BookingID    string    `json:"booking_id"`
}

func (r *BookingRepository) CreateBookingWithTickets(ctx context.Context, bookingParams CreateBookingParams, ticketParams []CreateTicketParams) (*Booking, []Ticket, error) {
	// เริ่ม transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	// สร้าง booking
	var booking Booking
	err = tx.QueryRowContext(ctx, `
		INSERT INTO bookings (status, total_price, expires_at, user_id, event_id)
		VALUES ('pending', $1, $2, $3, $4)
		RETURNING id, status, total_price, expires_at, created_at, updated_at, user_id, event_id
	`, bookingParams.TotalPrice, bookingParams.ExpiresAt, bookingParams.UserID, bookingParams.EventID,
	).Scan(
		&booking.ID, &booking.Status, &booking.TotalPrice, &booking.ExpiresAt,
		&booking.CreatedAt, &booking.UpdatedAt, &booking.UserID, &booking.EventID,
	)
	if err != nil {
		return nil, nil, err
	}

	// สร้าง tickets
	var tickets []Ticket
	for _, tp := range ticketParams {
		var ticket Ticket
		err = tx.QueryRowContext(ctx, `
			INSERT INTO tickets (status, ticket_pool_id, booking_id)
			VALUES ('reserved', $1, $2)
			RETURNING id, seat_number, status, created_at, updated_at, ticket_pool_id, booking_id
		`, tp.TicketPoolID, booking.ID,
		).Scan(
			&ticket.ID, &ticket.SeatNumber, &ticket.Status,
			&ticket.CreatedAt, &ticket.UpdatedAt, &ticket.TicketPoolID, &ticket.BookingID,
		)
		if err != nil {
			return nil, nil, err
		}
		tickets = append(tickets, ticket)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return &booking, tickets, nil
}

func (r *BookingRepository) GetBookingsByUserID(ctx context.Context, userID string) ([]Booking, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, status, total_price, expires_at, created_at, updated_at, user_id, event_id
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(
			&booking.ID, &booking.Status, &booking.TotalPrice, &booking.ExpiresAt,
			&booking.CreatedAt, &booking.UpdatedAt, &booking.UserID, &booking.EventID,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (r *BookingRepository) GetBookingByID(ctx context.Context, id string) (*Booking, error) {
	var booking Booking
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status, total_price, expires_at, created_at, updated_at, user_id, event_id
		FROM bookings
		WHERE id = $1
	`, id).Scan(
		&booking.ID, &booking.Status, &booking.TotalPrice, &booking.ExpiresAt,
		&booking.CreatedAt, &booking.UpdatedAt, &booking.UserID, &booking.EventID,
	)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) GetTicketPoolByID(ctx context.Context, id string) (*TicketPool, error) {
	var pool TicketPool
	err := r.db.QueryRowContext(ctx, `
		SELECT id, zone_name, price, total_capacity, available_capacity, display_name, created_at, updated_at, version, event_id
		FROM ticket_pools
		WHERE id = $1
	`, id).Scan(
		&pool.ID, &pool.ZoneName, &pool.Price, &pool.TotalCapacity,
		&pool.AvailableCapacity, &pool.DisplayName, &pool.CreatedAt,
		&pool.UpdatedAt, &pool.Version, &pool.EventID,
	)
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

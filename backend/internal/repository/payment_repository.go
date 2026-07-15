package repository

import (
	"context"
	"database/sql"
	"time"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

type Payment struct {
	ID                 string    `json:"id"`
	TransactionType    string    `json:"transaction_type"`
	Amount             float64   `json:"amount"`
	PaymentGatewayRef  string    `json:"payment_gateway_ref"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	BookingID          string    `json:"booking_id"`
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, bookingID string, amount float64) (*Payment, error) {
	// ใช้ transaction เพราะต้องอัปเดตหลาย table พร้อมกัน
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// สร้าง payment ledger
	var payment Payment
	err = tx.QueryRowContext(ctx, `
		INSERT INTO payment_ledgers (transaction_type, amount, payment_gateway_ref, status, booking_id)
		VALUES ('charge', $1, $2, 'success', $3)
		RETURNING id, transaction_type, amount, payment_gateway_ref, status, created_at, booking_id
	`, amount, "REF-"+bookingID[:8], bookingID,
	).Scan(
		&payment.ID, &payment.TransactionType, &payment.Amount,
		&payment.PaymentGatewayRef, &payment.Status, &payment.CreatedAt, &payment.BookingID,
	)
	if err != nil {
		return nil, err
	}

	// อัปเดต booking status → confirmed
	_, err = tx.ExecContext(ctx, `
		UPDATE bookings SET status = 'confirmed', updated_at = NOW()
		WHERE id = $1
	`, bookingID)
	if err != nil {
		return nil, err
	}

	// อัปเดต tickets status → confirmed
	_, err = tx.ExecContext(ctx, `
		UPDATE tickets SET status = 'confirmed', updated_at = NOW()
		WHERE booking_id = $1
	`, bookingID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetPaymentByBookingID(ctx context.Context, bookingID string) (*Payment, error) {
	var payment Payment
	err := r.db.QueryRowContext(ctx, `
		SELECT id, transaction_type, amount, payment_gateway_ref, status, created_at, booking_id
		FROM payment_ledgers
		WHERE booking_id = $1
	`, bookingID).Scan(
		&payment.ID, &payment.TransactionType, &payment.Amount,
		&payment.PaymentGatewayRef, &payment.Status, &payment.CreatedAt, &payment.BookingID,
	)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
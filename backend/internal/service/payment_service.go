package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
	"github.com/redis/go-redis/v9"
)

type PaymentService struct {
	repo        *repository.PaymentRepository
	bookingRepo *repository.BookingRepository
	rdb         *redis.Client
}

func NewPaymentService(
	repo *repository.PaymentRepository,
	bookingRepo *repository.BookingRepository,
	rdb *redis.Client,
) *PaymentService {
	return &PaymentService{repo: repo, bookingRepo: bookingRepo, rdb: rdb}
}

type CreatePaymentParams struct {
	BookingID string
	UserID    string
}

func (s *PaymentService) CreatePayment(ctx context.Context, params CreatePaymentParams) (*repository.Payment, error) {
	// ดึงข้อมูล booking
	booking, err := s.bookingRepo.GetBookingByID(ctx, params.BookingID)
	if err != nil {
		return nil, errors.New("ไม่พบ booking")
	}

	// เช็คว่าเป็น booking ของ user คนนี้
	if booking.UserID != params.UserID {
		return nil, errors.New("ไม่มีสิทธิ์ชำระเงิน booking นี้")
	}

	// เช็คว่า booking ยังไม่หมดอายุ
	if time.Now().After(booking.ExpiresAt) {
		return nil, errors.New("booking หมดอายุแล้ว")
	}

	// เช็คว่ายังไม่ได้ชำระเงิน
	if booking.Status != "pending" {
		return nil, fmt.Errorf("booking status เป็น '%s' ไม่สามารถชำระเงินได้", booking.Status)
	}

	// สร้าง payment
	payment, err := s.repo.CreatePayment(ctx, params.BookingID, booking.TotalPrice)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) GetPaymentByBookingID(ctx context.Context, bookingID string) (*repository.Payment, error) {
	return s.repo.GetPaymentByBookingID(ctx, bookingID)
}
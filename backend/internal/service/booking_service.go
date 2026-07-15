package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
	"github.com/redis/go-redis/v9"
)

type BookingService struct {
	repo *repository.BookingRepository
	rdb  *redis.Client
}

func NewBookingService(repo *repository.BookingRepository, rdb *redis.Client) *BookingService {
	return &BookingService{repo: repo, rdb: rdb}
}

type CreateBookingParams struct {
	UserID       string
	TicketPoolID string
	Quantity     int
}

type BookingResult struct {
	Booking *repository.Booking `json:"booking"`
	Tickets []repository.Ticket `json:"tickets"`
}

func (s *BookingService) CreateBooking(ctx context.Context, params CreateBookingParams) (*BookingResult, error) {
	// ดึงข้อมูล ticket pool
	pool, err := s.repo.GetTicketPoolByID(ctx, params.TicketPoolID)
	if err != nil {
		return nil, errors.New("ไม่พบ ticket pool")
	}

	// ใช้ Redis ลด available_capacity
	key := fmt.Sprintf("ticket_pool:%s:available", params.TicketPoolID)

	// ถ้ายังไม่มีใน Redis ให้ set ค่าจาก DB ก่อน
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		s.rdb.Set(ctx, key, pool.AvailableCapacity, 0)
	}

	// ลด capacity ตามจำนวนที่จอง
	remaining, err := s.rdb.DecrBy(ctx, key, int64(params.Quantity)).Result()
	if err != nil {
		return nil, err
	}

	// ถ้าเหลือน้อยกว่า 0 แปลว่าตั๋วหมด
	if remaining < 0 {
		s.rdb.IncrBy(ctx, key, int64(params.Quantity)) // คืนค่ากลับ
		return nil, errors.New("ตั๋วไม่เพียงพอ")
	}

	// คำนวณราคารวม
	totalPrice := pool.Price * float64(params.Quantity)

	// สร้าง ticket params
	var ticketParams []repository.CreateTicketParams
	for i := 0; i < params.Quantity; i++ {
		ticketParams = append(ticketParams, repository.CreateTicketParams{
			TicketPoolID: params.TicketPoolID,
		})
	}

	// สร้าง booking + tickets ใน DB
	booking, tickets, err := s.repo.CreateBookingWithTickets(ctx,
		repository.CreateBookingParams{
			UserID:     params.UserID,
			EventID:    pool.EventID,
			TotalPrice: totalPrice,
			ExpiresAt:  time.Now().Add(15 * time.Minute),
		},
		ticketParams,
	)
	if err != nil {
		// ถ้า DB ล้มเหลว คืน capacity กลับ Redis
		s.rdb.IncrBy(ctx, key, int64(params.Quantity))
		return nil, err
	}

	return &BookingResult{
		Booking: booking,
		Tickets: tickets,
	}, nil
}

func (s *BookingService) GetBookingsByUserID(ctx context.Context, userID string) ([]repository.Booking, error) {
	return s.repo.GetBookingsByUserID(ctx, userID)
}

func (s *BookingService) GetBookingByID(ctx context.Context, id string) (*repository.Booking, error) {
	return s.repo.GetBookingByID(ctx, id)
}

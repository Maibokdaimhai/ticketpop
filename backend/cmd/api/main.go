package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Maibokdaimhai/ticketpop/internal/handler"
	"github.com/Maibokdaimhai/ticketpop/internal/repository"
	"github.com/Maibokdaimhai/ticketpop/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("ไม่พบไฟล์ .env:", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("ไม่สามารถเชื่อมต่อ database ได้:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("database ไม่ตอบสนอง:", err)
	}
	log.Println("เชื่อมต่อ database สำเร็จ!")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT"),
		),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis ไม่ตอบสนอง:", err)
	}
	log.Println("เชื่อมต่อ Redis สำเร็จ!")

	// user
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// event
	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)

	// ticketPool
	ticketPoolRepo := repository.NewTicketPoolRepository(db)
	ticketPoolService := service.NewTicketPoolService(ticketPoolRepo)
	ticketPoolHandler := handler.NewTicketPoolHandler(ticketPoolService)

	// booking
	bookingRepo := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepo, rdb)
	bookingHandler := handler.NewBookingHandler(bookingService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// Event routes
	events := r.Group("/events")
	{
		events.GET("", eventHandler.GetEvents)
		events.GET("/:id", eventHandler.GetEventByID)
		events.GET("/:id/ticket-pools", ticketPoolHandler.GetTicketPoolsByEventID)
	}

	// Protected routes (ต้อง login)
	api := r.Group("/api")
	api.Use(handler.AuthMiddleware())
	{
		api.GET("/me", userHandler.Me)
		api.POST("/events", eventHandler.CreateEvent)
		api.POST("/events/:id/ticket-pools", ticketPoolHandler.CreateTicketPool)
		api.POST("/bookings", bookingHandler.CreateBooking)
		api.GET("/bookings", bookingHandler.GetBookings)
		api.GET("/bookings/:id", bookingHandler.GetBookingByID)
	}

	r.Run(":8080")
}

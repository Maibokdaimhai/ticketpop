package main

import (
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

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

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
	// Protected routes (ต้อง login)
	api := r.Group("/api")
	api.Use(handler.AuthMiddleware())
	{
		api.GET("/me", userHandler.Me)
	}

	r.Run(":8080")
}

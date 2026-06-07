package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    // โหลด .env file
    if err := godotenv.Load(); err != nil {
        log.Fatal("ไม่พบไฟล์ .env:", err)
    }

    // อ่านค่าจาก environment variable
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

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
        })
    })

    r.Run(":8080")
}
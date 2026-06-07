package main

import (
    "database/sql"
    "log"

    "github.com/gin-gonic/gin"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    // เชื่อมต่อ PostgreSQL
    db, err := sql.Open("pgx", "postgres://ticketpop:password@localhost:5432/ticketpop")
    if err != nil {
        log.Fatal("ไม่สามารถเชื่อมต่อ database ได้:", err)
    }
    defer db.Close()

    // ทดสอบว่าเชื่อมได้จริงๆ
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
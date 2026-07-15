package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึง token จาก Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "กรุณา login ก่อน",
			})
			c.Abort()
			return
		}

		// header ต้องเป็นแบบ "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "รูปแบบ token ไม่ถูกต้อง",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// ตรวจสอบ token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token ไม่ถูกต้องหรือหมดอายุ",
			})
			c.Abort()
			return
		}

		// ดึง user_id จาก token แล้วเก็บไว้ใน context
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])

		c.Next()
	}
}

func RoleMiddleware(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "ไม่มีสิทธิ์เข้าถึง",
            })
            c.Abort()
            return
        }

        // เช็คว่า role ของ user อยู่ใน roles ที่อนุญาตไหม
        userRole := role.(string)
        for _, r := range roles {
            if userRole == r {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{
            "error": "ไม่มีสิทธิ์เข้าถึง",
        })
        c.Abort()
    }
}
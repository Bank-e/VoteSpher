package middleware

import (
	"net/http" // net/http สำหรับ Status Code
	"os"
	"strings"

	"votespher/pkg"

	"github.com/gin-gonic/gin"
)

// 1. ด่านตรวจ Token
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error_code": "UNAUTHORIZED",
				"message":    "กรุณา login ก่อน",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkg.ValidateToken(tokenStr, os.Getenv("JWT_SECRET_KEY"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message":    "token ไม่ถูกต้องหรือหมดอายุ",
			})
			return
		}

		// เก็บ claims ลง Gin context แทน r.Context()
		c.Set("role", claims.Role)
		c.Set("voter_id", claims.VoterID)
		c.Set("area_id", claims.AreaID)
		c.Next()
	}
}

// 2. ด่านตรวจสิทธิ์
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get("role")
		
		// แปลงเป็นตัวเล็กทั้งคู่เพื่อป้องกัน Error จากพิมพ์เล็ก/ใหญ่
		if !strings.EqualFold(userRole.(string), requiredRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message":    "ไม่มีสิทธิ์เข้าถึง",
			})
			return
		}
		c.Next()
	}
}
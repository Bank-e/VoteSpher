package middleware

import (
	"net/http" // net/http สำหรับ Status Code
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"votespher/pkg"
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
				"error_code": "UNAUTHORIZED",
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
		if userRole != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error_code": "FORBIDDEN",
				"message":    "ไม่มีสิทธิ์เข้าถึง",
			})
			return
		}
		c.Next()
	}
}
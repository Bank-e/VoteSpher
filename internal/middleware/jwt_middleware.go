package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"votespher/pkg"
)

// 1. ด่านตรวจ Token
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "กรุณา login ก่อน", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkg.ValidateToken(tokenStr, os.Getenv("JWT_SECRET_KEY"))
		if err != nil {
			http.Error(w, "token ไม่ถูกต้องหรือหมดอายุ", http.StatusUnauthorized)
			return
		}

		// ฝังข้อมูลทั้งหมดที่จำเป็นลงใน Context เพื่อให้ Handler ตัวต่อไปใช้ได้
		ctx := context.WithValue(r.Context(), "role", claims.Role)
		ctx = context.WithValue(ctx, "voter_id", claims.VoterID)
		ctx = context.WithValue(ctx, "area_id", claims.AreaID)
		
		// เรียกใช้ Handler ตัวต่อไป (หรือ Middleware ตัวต่อไป) พร้อมส่ง Context ใหม่ไป
		next(w, r.WithContext(ctx))
	}
}

// 2. ด่านตรวจสิทธิ์
func RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ดึง role ที่ RequireAuth ฝังไว้
		userRole := r.Context().Value("role")

		// ถ้าสิทธิ์ไม่ตรง ให้เตะออก
		if userRole != requiredRole {
			http.Error(w, "ไม่มีสิทธิ์เข้าถึง", http.StatusForbidden)
			return
		}

		// ถ้าสิทธิ์ตรง ให้ไปต่อที่ Handler หลัก
		next(w, r)
	}
}
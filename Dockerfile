# ==========================================
# Stage 1: Builder (สร้างสภาพแวดล้อมสำหรับ Build โค้ด)
# ==========================================
FROM golang:1.26-alpine AS builder

# ตั้งค่า Working Directory
WORKDIR /app

# Copy ไฟล์จัดการ Package มาก่อนเพื่อทำ Caching (ช่วยให้ Build รอบหลังเร็วขึ้น)
COPY go.mod go.sum ./
RUN go mod download

# Copy โค้ดทั้งหมดเข้ามา
COPY . .

# สั่ง Build โค้ด Go ให้เป็น Binary ไฟล์เดียว
# - CGO_ENABLED=0: ปิด CGO เพื่อให้ได้ไฟล์ Static binary แท้ๆ รันได้ทุกที่
# - GOOS=linux: บังคับว่าให้รันบน Linux
# - o votespher-app: ตั้งชื่อไฟล์ที่ได้ว่า votespher-app
RUN CGO_ENABLED=0 GOOS=linux go build -o votespher-app ./cmd/server/main.go


# ==========================================
# Stage 2: Runner (สร้าง Image ตัวจริงสำหรับใช้งาน)
# ==========================================
FROM alpine:latest

WORKDIR /app

# ติดตั้ง Certificate สำหรับต่อ Cloud DB และตั้งค่า Timezone ให้ตรงกับไทย (แก้ปัญหาเวลาเพี้ยน)
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Bangkok

# Copy ไฟล์ Binary ที่ Build เสร็จแล้วจาก Stage 1 มาใส่ใน Stage 2
COPY --from=builder /app/votespher-app .

# (ตัวเลือก) ถ้าโหมด Cloud ของคุณต้องใช้ไฟล์ ca.pem ให้ copy มาด้วย
# COPY ca.pem ./ca.pem

# เปิด Port 8080 (หรือพอร์ตที่คุณตั้งไว้ใน Gin/Fiber)
EXPOSE 8080

# คำสั่งสำหรับรันแอปพลิเคชัน
CMD ["./votespher-app"]
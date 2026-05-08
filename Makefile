 # A `Makefile` is a file used to define commands for building, running, testing, and managing a project using the make command.
# Instead of typing long commands, you run:
# make run
# make build
# make tidy

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

tidy:
	go mod tidy

# รันเทสทั้งหมดแบบรวดเร็ว
test:
	go test ./...

# รันเทสทั้งหมดแบบดูรายละเอียด (บอกว่าผ่าน/ไม่ผ่าน ทีละฟังก์ชัน)
test-v:
	go test -v ./...

# รันเทสและดูเปอร์เซ็นต์ Coverage ใน Terminal
test-cover:
	go test -cover ./...

# รันเทส สร้างไฟล์สถิติ และเปิดหน้าเว็บ HTML ให้ดูแบบสวยงาม
test-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	@echo "🌐 เปิดหน้าเว็บแสดงผล Coverage เรียบร้อยแล้ว!"

# ลบไฟล์ขยะที่เกิดจากการเทส (ทำความสะอาดโปรเจกต์)
test-clean:
	-@del /f coverage.out 2>nul || rm -f coverage.out


# ==============================================================================
# VoteSpher Docker Commands (API & Database)
# ==============================================================================

# สั่งรันทั้งระบบ (API และ DB) พร้อมสั่ง Build โค้ด Go ใหม่เสมอ
up:
	@echo "🚀 Starting VoteSpher Application..."
	docker-compose up -d --build
	docker ps

# สั่งหยุดชั่วคราว (คอนเทนเนอร์ยังอยู่ครบ)
stop:
	@echo "⏸️ Stopping all services..."
	docker-compose stop

# สั่งปิดและลบคอนเทนเนอร์ (ข้อมูล Database ยังอยู่)
down:
	@echo "🛑 Stopping and removing all containers..."
	docker-compose down

# ⚠️ สั่งล้างบาง! ลบคอนเทนเนอร์และลบข้อมูล Database ทิ้งทั้งหมด (ใช้ตอน Reset ระบบ)
clean:
	@echo "🧹 WARNING: Removing containers AND clearing all data..."
	docker-compose down -v

# ==============================================================================
# Logs Commands (กด Ctrl+C เพื่อออก)
# ==============================================================================

# ดู Log รวมทั้ง API และ DB วิ่งพร้อมกัน
logs:
	docker-compose logs -f

# ดู Log เฉพาะของ API อย่างเดียว (เอาไว้ดูตอนยิง Request หรือ Debug โค้ด Go)
logs-api:
	docker-compose logs -f api

# ดู Log เฉพาะของ Database อย่างเดียว
logs-db:
	docker-compose logs -f db
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

# รันเทสเฉพาะระบบ Voting ที่เพิ่งทำ Clean Architecture
test-voting:
# 	go test -v ./internal/voting
	go test -v -coverprofile=voting_coverage.out ./internal/voting && go tool cover -html=voting_coverage.out

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
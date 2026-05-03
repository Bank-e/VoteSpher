# File structure

---

```File_structure

VOTESPHER/                           # root ของโปรเจกต์ทั้งหมด
│
├── cmd/                             # entry point ของแอปพลิเคชัน
│   └── server/                      # HTTP server หลัก
│       └── main.go                  # จุดเริ่มต้นโปรแกรม: เริ่ม server, โหลด config,
│                                    # register routes, inject dependencies
│
├── config/                          # ตั้งค่าการเชื่อมต่อและ environment ต่าง ๆ
│   └── database.go                  # เชื่อมต่อฐานข้อมูล (DSN, connection pool,
│                                    # ping check) อ่านค่าจาก env vars
│
├── internal/                        # business logic หลักทั้งหมด แบ่งตาม domain
│   │                                # แต่ละ module ใช้ pattern: Handler → Service → Repository
│   │
│   ├── auth/                        # จัดการ authentication และ authorization
│   │   ├── handler.go               # รับ HTTP request: login, logout, ขอ OTP
│   │   │                            # ส่ง response JSON กลับ client
│   │   ├── model.go                 # struct: LoginRequest, LoginResponse,
│   │   │                            # OTPRequest, TokenClaims
│   │   ├── repository.go            # query ฐานข้อมูล: ค้นหา voter, บันทึก OTP,
│   │   │                            # ตรวจสอบ session
│   │   └── service.go               # business logic: ตรวจ OTP, สร้าง JWT token,
│   │                                # validate credentials
│   │
│   ├── info/                        # ดึงข้อมูลสำหรับแสดงผล (read-only mostly)
│   │   ├── handler.go               # endpoint: GET candidates, GET parties, GET system config
│   │   ├── model.go                 # struct: Candidate, Party, SystemConfig
│   │   ├── repository.go            # query ดึงรายชื่อผู้สมัคร, พรรค, ข้อมูลการเลือกตั้ง
│   │   └── service.go               # จัดรูปแบบข้อมูล, กรองตามเงื่อนไข
│   │
│   ├── middleware/                  # middleware ที่รันก่อน handler ทุก request
│   │   └── jwt_middleware.go        # ตรวจสอบ JWT token ใน Authorization header
│   │                                # ถ้า token ไม่ valid → ตอบ 401 Unauthorized
│   │                                # ถ้าผ่าน → ส่ง claims ต่อให้ handler ผ่าน context
│   │
│   ├── result/                      # คำนวณและแสดงผลการเลือกตั้ง
│   │   ├── handler.go               # endpoint: GET /results — ผลรวมคะแนนทั้งหมด
│   │   ├── model.go                 # struct: VoteResult, CandidateScore, Summary
│   │   ├── repository.go            # query นับคะแนน GROUP BY candidate
│   │   └── service.go               # คำนวณ % คะแนน, จัดอันดับ, สรุปผล
│   │
│   └── voting/                      # จัดการการโหวตของผู้ใช้
│       ├── handler.go               # endpoint: POST /vote — รับคะแนนจาก voter
│       ├── model.go                 # struct: VoteRequest, VoteRecord
│       ├── repository.go            # INSERT vote, ตรวจว่า voter โหวตไปแล้วหรือยัง
│       └── service.go               # validate: ช่วงเวลา, สิทธิ์, ซ้ำซ้อน
│                                    # เรียก repository เพื่อบันทึก vote
│
├── pkg/                             # shared utilities ที่ใช้ได้ทั่วทั้งโปรเจกต์
│   └── jwt.go                       # helper functions: GenerateToken(claims),
│                                    # ParseToken(tokenStr) → คืน claims หรือ error
│
├── go.mod                           # จัดการ dependencies (module name, Go version, packages)
├── Makefile                         # คำสั่งลัด: make run / make migrate / make build
├── File structure.md                # เอกสารอธิบายโครงสร้างโปรเจกต์
└── README.md                        # คู่มือการติดตั้งและใช้งานโปรเจกต์

# ============================================================
# DATA FLOW
# ============================================================
#
#   HTTP Request
#       ↓
#   JWT Middleware
#       ↓
#   Handler         ← รับ request, parse body/params
#       ↓
#   Service         ← business logic, validation
#       ↓
#   Repository      ← query / write ฐานข้อมูล
#       ↓
#   Database (PostgreSQL / MySQL หรืออื่นๆ)
#
# ============================================================
```

# File structure NEW

---

```structure
VOTESPHER/
├── cmd/
│   └── server/
│       └── main.go           # Entry point: จุดเริ่มต้นโปรแกรมและการทำ Dependency Injection
├── config/
│   └── database.go         # Configuration: จัดการการเชื่อมต่อฐานข้อมูล (PostgreSQL/Redis)
├── internal/               # Private application and library code
│   ├── auth/               # Module: ระบบยืนยันตัวตนและการจัดการสิทธิ์
│   ├── election/           # Module: การจัดการข้อมูลการเลือกตั้ง
│   ├── info/               # Module: การดึงข้อมูลทั่วไป (Candidates, Parties, Config)
│   ├── result/             # Module: ระบบคำนวณและประมวลผลคะแนน
│   ├── voting/             # Module: ระบบลงคะแนน (Voting Logic)
│   ├── middleware/         # Custom Middleware (e.g., JWT Authentication)
│   └── models/             # Global Domain Entities (Database Models)
├── migration/
│   ├── migrate.go          # Database Migration (Schema definition)
│   └── seed.go             # Data Seeding (Initial setup data)
├── pkg/                    # Shared utilities (e.g., JWT helper, Hashing)
├── .env                    # Environment variables (Sensitive data)
├── ca.pem                  # SSL/TLS Certificate สำหรับการเชื่อมต่อฐานข้อมูลแบบ Secure (e.g., Cloud DB)
├── go.mod                  # Go Modules: ระบุชื่อโปรเจกต์และเวอร์ชันของ Library ทั้งหมดที่ใช้
├── go.sum                  # Checksums: ไฟล์ที่ใช้ยืนยันความถูกต้องและปลอดภัยของ Dependencies
├── Makefile                # Shortcut commands (make run, make migrate)
└── README.md
```

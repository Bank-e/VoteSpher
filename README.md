# VoteSpher — ระบบเลือกตั้งออนไลน์

ระบบเลือกตั้งออนไลน์พร้อม Two-Factor Authentication (บัตรประชาชน + OTP ทางอีเมล), Secret Ballot, และ Real-time Results Dashboard

**Stack:** Go 1.26 / Gin / GORM / MySQL · React 18 / Vite / Tailwind CSS

---

## สิ่งที่ต้องติดตั้ง

| Tool | เวอร์ชันขั้นต่ำ |
|------|--------------|
| Go | 1.22+ |
| Node.js | 18+ |
| npm | 9+ |
| MySQL | 5.7+ (หรือ Aiven Cloud) |

---

## การติดตั้ง

### 1. Clone โปรเจกต์

```bash
git clone https://github.com/Xagatech/VoteSphertestfrontend.git
cd VoteSphertestfrontend
git checkout develop
```

### 2. ตั้งค่า Environment Variables

คัดลอกไฟล์ตัวอย่างและแก้ไขค่าให้ตรงกับระบบของคุณ

```bash
cp .env.example .env
```

แก้ไข `.env`:

```env
# ─── Database ───────────────────────────────────────
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=votespher
DB_CA_CERT=               # ใส่ path ถ้าใช้ TLS (Aiven ใส่ ca.pem)

# ─── Authentication ──────────────────────────────────
JWT_SECRET_KEY=your_jwt_secret_key_here
HASH_SECRET_KEY=your_hash_secret_key_here
JWT_EXPIRY_HOURS=2

# ─── Email (Gmail SMTP) ──────────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password   # Gmail App Password (ไม่ใช่รหัสผ่านปกติ)
SMTP_FROM=your_email@gmail.com

# ─── Server ──────────────────────────────────────────
PORT=8080
CORS_ALLOWED_ORIGIN=http://localhost:3000

# ─── Dev Mode ────────────────────────────────────────
ENABLE_DEV_ENDPOINTS=false   # true เพื่อเปิด /dev/mock-token
```

> **Gmail App Password:** Google Account → Security → 2-Step Verification → App passwords

### 3. ติดตั้ง Go Dependencies

```bash
go mod download
```

### 4. สร้างฐานข้อมูล

```bash
# สร้าง Schema (ครั้งแรกเท่านั้น)
RUN_MIGRATION=true go run cmd/server/main.go

# ใส่ข้อมูลตัวอย่าง (optional)
RUN_SEED=true go run cmd/server/main.go
```

### 5. ติดตั้ง Frontend Dependencies

```bash
cd frontend
npm install
cd ..
```

---

## การรันระบบ (Development)

เปิด **2 terminal** พร้อมกัน:

**Terminal 1 — Backend:**
```bash
go run cmd/server/main.go
# Server รันที่ http://localhost:8080
```

**Terminal 2 — Frontend:**
```bash
cd frontend
npm run dev
# Frontend รันที่ http://localhost:3000
```

เปิด browser ไปที่ **http://localhost:3000**

---

## การใช้งาน

### ขั้นตอน Login

1. กรอกเลขบัตรประชาชน 13 หลัก (ไม่มีขีด)
2. ระบบส่ง OTP 6 หลักไปยังอีเมลที่ลงทะเบียนไว้
3. กรอก OTP เพื่อรับ JWT Token

### Dev Mode (ทดสอบโดยไม่ต้องรับอีเมลจริง)

ตั้ง `ENABLE_DEV_ENDPOINTS=true` ใน `.env` แล้ว restart backend

```
citizen_id : 1111111111111
OTP        : 111111   (OTP ตายตัวในโหมด dev)
```

### บทบาทผู้ใช้

| Role | สิ่งที่ทำได้ |
|------|-------------|
| **Voter** | Login → เลือก Candidate → ลงคะแนน → ดูผล |
| **Admin** | ทุกอย่างของ Voter + ควบคุม Election State |

### Election State Machine

```
PREPARE → OPEN → PAUSED → OPEN → CLOSED → COUNTING
```

Admin เปลี่ยน State ได้ที่หน้า Admin Panel (ต้อง login ด้วย account ที่เป็น admin)

---

## API Endpoints

### Public
| Method | Path | คำอธิบาย |
|--------|------|---------|
| POST | `/voter/verify` | ตรวจสอบบัตรประชาชน |
| POST | `/voter/otp-request` | ขอ OTP |
| POST | `/voter/otp-confirm` | ยืนยัน OTP รับ JWT |
| GET | `/candidates?area_id=1` | รายชื่อผู้สมัครในเขต |
| GET | `/parties` | รายชื่อพรรคทั้งหมด |
| GET | `/election/config` | สถานะการเลือกตั้งปัจจุบัน |
| GET | `/results/areas` | ผลโหวตรวมทุกเขต |
| GET | `/results/areas/:id` | ผลโหวตแยกพรรคในเขต |

### Protected (ต้องใช้ JWT)
| Method | Path | คำอธิบาย |
|--------|------|---------|
| GET | `/voter/me` | ข้อมูลตัวเอง |
| POST | `/ballot/submit` | ส่งคะแนนโหวต |
| GET | `/ballot/status` | เช็คว่าโหวตแล้วหรือยัง |

### Admin (ต้องใช้ JWT + role=admin)
| Method | Path | คำอธิบาย |
|--------|------|---------|
| PATCH | `/election/config` | แก้ไขสถานะการเลือกตั้ง |

### Dev (เฉพาะ ENABLE_DEV_ENDPOINTS=true)
| Method | Path | คำอธิบาย |
|--------|------|---------|
| POST | `/dev/mock-token` | สร้าง JWT สมมติโดยไม่ต้อง login |

```bash
curl -X POST http://localhost:8080/dev/mock-token \
  -H "Content-Type: application/json" \
  -d '{"voter_id": 1, "area_id": 1, "role": "admin"}'
```

---

## Build สำหรับ Production

### Backend

```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

### Frontend

```bash
cd frontend
npm run build
# Output อยู่ที่ frontend/dist/
```

### Docker

```bash
docker build -t votespher .
docker run -p 8080:8080 --env-file .env votespher
```

---

## Makefile Commands

```bash
make run          # รัน backend dev server
make build        # build binary
make test         # รัน unit tests ทั้งหมด
make test-voting  # รัน tests + coverage สำหรับ voting module
make test-html    # เปิด coverage report ใน browser
```

---

## โครงสร้างโปรเจกต์

```
VoteSpher/
├── cmd/server/main.go          # Entry point
├── config/                     # Database connection
├── internal/
│   ├── auth/                   # Authentication (verify, OTP, JWT)
│   ├── election/               # Election state machine
│   ├── info/                   # Candidates & parties (read-only)
│   ├── middleware/             # JWT, rate limit, audit log
│   ├── models/                 # GORM models
│   ├── realtime/               # Live vote aggregation
│   ├── result/                 # Vote result queries
│   └── voting/                 # Core ballot submission
├── pkg/                        # JWT, Email, Email Queue
├── migration/                  # Schema migration & seed data
├── frontend/
│   ├── src/
│   │   ├── App.jsx
│   │   ├── pages/
│   │   │   ├── VotePage.jsx
│   │   │   ├── ResultsPage.jsx
│   │   │   └── AdminPage.jsx
│   │   └── lib/api.js
│   └── public/
│       └── images/parties/     # โลโก้พรรคการเมือง (local)
├── .env.example
├── Dockerfile
├── Makefile
└── go.mod
```

---

## ความปลอดภัย

- **Secret Ballot** — Vote record ไม่เก็บ voter_id เด็ดขาด
- **Citizen ID Hashing** — เก็บแค่ HMAC-SHA256 ไม่เก็บข้อมูลดิบ
- **OTP Lockout** — กรอกผิด 5 ครั้ง OTP ถูกยกเลิกทันที
- **Row-level Locking** — ป้องกัน duplicate vote จาก concurrent requests
- **Rate Limiting** — 60 requests/minute/IP
- **JWT Authentication** — HS256, expiry configurable
- **Audit Log** — บันทึกทุก ballot submission พร้อม IP และ timestamp

---

## Deploy บน Railway + Vercel

**Backend (Railway):**
1. Connect repository ใน Railway Dashboard
2. ตั้ง Environment Variables ใน Railway
3. Railway inject `PORT` ให้อัตโนมัติ

**Frontend (Vercel):**
1. Connect repository ใน Vercel
2. Set Root Directory เป็น `frontend`
3. ตั้ง `VITE_API_URL` = URL ของ Railway backend

---

## License

MIT

# VoteSpher — ระบบเลือกตั้งออนไลน์

ระบบเลือกตั้งออนไลน์พร้อม Two-Factor Authentication (บัตรประชาชน + OTP ทาง Email/SMS), Secret Ballot, และ Real-time Results

**Stack:** Go 1.23 / Gin / GORM / MySQL · React 18 / Vite / Tailwind CSS

---

## สิ่งที่ต้องติดตั้ง

| Tool | เวอร์ชันขั้นต่ำ |
|------|--------------|
| Go | 1.22+ |
| Node.js | 18+ |
| MySQL | 5.7+ |

---

## ติดตั้งและรันบนเครื่อง (Local)

### 1. Clone โปรเจกต์

```bash
git clone https://github.com/Xagatech/vote.git
cd vote
```

### 2. ตั้งค่า Environment Variables

```bash
cp .env.example .env
```

แก้ไข `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=votespher

# Auth
JWT_SECRET_KEY=change_this_to_random_string
HASH_SECRET_KEY=change_this_to_another_random_string
JWT_EXPIRY_HOURS=2

# Email (Gmail SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_gmail_app_password
SMTP_FROM=your_email@gmail.com

# Server
PORT=8080
CORS_ALLOWED_ORIGIN=http://localhost:3000

# Dev mode — เปิดเพื่อใช้ OTP 111111 และ /dev/mock-token
ENABLE_DEV_ENDPOINTS=false
```

> **Gmail App Password:** Google Account → Security → 2-Step Verification → App passwords

### 3. ติดตั้ง Dependencies

```bash
# Backend
go mod download

# Frontend
cd frontend && npm install && cd ..
```

### 4. สร้างฐานข้อมูล

```bash
# สร้าง schema (ครั้งแรก)
RUN_MIGRATION=true go run cmd/server/main.go

# เพิ่มข้อมูลตัวอย่าง (optional)
RUN_SEED=true go run cmd/server/main.go
```

### 5. รันระบบ

เปิด **2 terminal**:

```bash
# Terminal 1 — Backend (http://localhost:8080)
go run cmd/server/main.go

# Terminal 2 — Frontend (http://localhost:3000)
cd frontend && npm run dev
```

เปิด **http://localhost:3000**

---

## วิธีใช้งาน

### ขั้นตอน Login

1. กรอกเลขบัตรประชาชน 13 หลัก
2. เลือกช่องทางรับ OTP (Email หรือ SMS)
3. กรอก OTP 6 หลักที่ได้รับ
4. ระบบพาไปหน้าโหวต (voter) หรือ Admin Dashboard (admin)

### บทบาทผู้ใช้

| Role | สิ่งที่ทำได้ |
|------|-------------|
| **Voter** | Login → เลือก Candidate → ลงคะแนน → ดูผล Realtime |
| **Admin** | ทุกอย่างของ Voter + ควบคุม Election State |

### Election State Machine

```
PREPARE → OPEN → PAUSED → OPEN → CLOSED → COUNTING
```

Voter โหวตได้เฉพาะตอน state = **OPEN**

---

## Dev Mode

ตั้ง `ENABLE_DEV_ENDPOINTS=true` แล้ว restart backend

- OTP ตายตัว: **111111**
- Bypass login: `POST /dev/mock-token`

```bash
curl -X POST http://localhost:8080/dev/mock-token \
  -H "Content-Type: application/json" \
  -d '{"voter_id": 21, "area_id": 1, "role": "voter"}'
```

Test accounts ดูได้ที่ `DEMO.md`

---

## Makefile

```bash
make run        # รัน backend dev server
make build      # build binary → bin/server
make test       # รัน unit tests ทั้งหมด
```

---

## โครงสร้างโปรเจกต์

```
vote/
├── cmd/server/main.go       # Entry point
├── config/                  # Database connection
├── internal/
│   ├── auth/                # Verify, OTP, JWT
│   ├── election/            # Election state machine
│   ├── info/                # Candidates & parties
│   ├── middleware/          # JWT, rate limit, audit log
│   ├── realtime/            # Live vote aggregation
│   ├── result/              # Vote result queries
│   └── voting/              # Ballot submission
├── pkg/                     # JWT, Email, SMS, Queue
├── migration/               # Schema + seed data
├── frontend/                # React 18 / Vite / Tailwind
├── .env.example
├── Dockerfile
├── railway.toml             # Railway deploy config
└── Makefile
```

---

## ความปลอดภัย

- **Secret Ballot** — ไม่เชื่อม voter กับ candidate ที่เลือกในฐานข้อมูล
- **Citizen ID Hashing** — เก็บแค่ HMAC-SHA256
- **OTP Lockout** — กรอกผิด 5 ครั้งล็อกทันที
- **Row-level Locking** — ป้องกัน duplicate vote จาก concurrent requests
- **Rate Limiting** — 60 req/min/IP
- **Audit Log** — บันทึกทุก ballot submission

---

## Deploy Online

### Backend → Railway

1. เข้า [railway.app](https://railway.app) → New Project → Deploy from GitHub
2. เลือก repo นี้
3. ตั้ง Environment Variables ตาม `.env.example`
4. Railway ใช้ `Dockerfile` + `railway.toml` อัตโนมัติ

### Frontend → Vercel

1. เข้า [vercel.com](https://vercel.com) → New Project → Import repo นี้
2. Set **Root Directory** = `frontend`
3. ตั้ง Environment Variable: `VITE_API_URL` = URL ของ Railway backend
4. Deploy

### Database → Aiven (MySQL Cloud)

1. สร้าง MySQL service ที่ [aiven.io](https://aiven.io)
2. ใส่ connection string ใน Railway environment variables
3. ถ้าใช้ TLS ให้ download `ca.pem` แล้วตั้ง `DB_CA_CERT` path

---

## License

MIT

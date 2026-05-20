# VoteSpher — ระบบเลือกตั้งออนไลน์

ระบบเลือกตั้งออนไลน์พร้อม Two-Factor Authentication (บัตรประชาชน + OTP ทาง Email/SMS), Secret Ballot, และ Real-time Results

**Stack:** Go 1.23 / Gin / GORM / MySQL · React 18 / Vite / Tailwind CSS

---

## สิ่งที่ต้องติดตั้ง

| Tool | เวอร์ชันขั้นต่ำ |
|------|--------------|
| Go | 1.22+ |
| Node.js | 18+ |
| Docker | 24+ (สำหรับการรันด้วย Docker Compose) |
| MySQL | 5.7+ (กรณีไม่ได้รันด้วย Docker) |

---

## 🚀 การติดตั้งและรันระบบ (วิธีที่แนะนำ)

การรันด้วย Docker Compose จะทำให้คุณไม่ต้องติดตั้ง Database หรือตั้งค่าอะไรให้วุ่นวาย เพราะระบบจะรันทุกอย่าง (DB, API, Frontend) ขึ้นมาให้พร้อมใช้งาน

### 1. Clone โปรเจกต์ และตั้งค่า Environment

```bash
git clone https://github.com/Xagatech/vote.git
cd vote
cp .env.example .env
```
*(แก้ไขไฟล์ `.env` ตามความเหมาะสม ดูคำอธิบายตัวแปรได้ในไฟล์)*

### 2. รันระบบด้วย Docker Compose

เราเตรียมคำสั่งลัดไว้ให้ใน `Makefile` แล้ว:

```bash
# สั่งรันทั้งระบบ (API, DB, Frontend) จะใช้เวลา Build ครั้งแรกสักครู่
make up
```

**รอจนกว่า Container ทั้งหมดจะพร้อมใช้งาน สามารถเข้าเว็บไซต์ได้เลยที่:**
- **Frontend App:** [http://localhost:3000](http://localhost:3000)
- **Backend API:** [http://localhost:8080](http://localhost:8080)

หากต้องการหยุดระบบ:
```bash
make down      # ปิดระบบ (ข้อมูลยังอยู่)
make stop      # หยุดชั่วคราว
make clean     # ⚠️ ลบระบบและเคลียร์ข้อมูล Database ทิ้งทั้งหมด
```

---

## 💻 การรันแบบ Local (สำหรับ Development)

หากคุณต้องการเขียนโค้ดและรันแบบ Local ล้วนๆ (Hot-reload):

**1. ติดตั้ง Dependencies**
```bash
go mod download
cd frontend && npm install && cd ..
```

**2. รัน Backend (Terminal 1)**
```bash
# ถ้าต้องการสร้างตารางหรือใส่ข้อมูลเริ่มต้น ให้เพิ่ม RUN_MIGRATION=true และ RUN_SEED=true นำหน้า
go run cmd/server/main.go
```

**3. รัน Frontend (Terminal 2)**
```bash
cd frontend && npm run dev
```

เปิด **http://localhost:3000** ใช้งานได้ทันที (Vite จะทำ Proxy ส่ง API ไปหา Backend ที่พอร์ต 8080 ให้อัตโนมัติ)

---

## 📁 โครงสร้างโปรเจกต์ (Project Structure)

```
vote/
├── api/                     # สเปก API (openapi.yaml) สำหรับ Swagger
├── cmd/server/main.go       # Entry point ฝั่ง Backend
├── config/                  # Database connection config
├── frontend/                # React 18 / Vite / Tailwind
│   ├── nginx.conf           # Reverse Proxy Config สำหรับ Production
│   └── vercel.json          # Config สำหรับ Deploy บน Vercel
├── internal/
│   ├── auth/                # Verify, OTP, JWT
│   ├── election/            # Election state machine
│   ├── info/                # Candidates & parties
│   ├── middleware/          # JWT, rate limit, audit log
│   ├── models/              # Global Domain Entities (Database Models)
│   ├── realtime/            # Live vote aggregation
│   ├── result/              # Vote result queries
│   └── voting/              # Ballot submission
├── migration/               # Schema + seed data
├── pkg/                     # JWT, Email, SMS, Queue
├── postman/                 # Postman Collection สำหรับเทส API
├── .env.example
├── Dockerfile.backend       # Docker สำหรับ Go API
├── Dockerfile.frontend      # Docker สำหรับ React App (ใช้ Nginx Proxy)
├── docker-compose.yml       # รันทั้งระบบ
└── Makefile                 # คำสั่งลัดการทำงาน
```

---

## 🛠️ คำสั่ง Makefile อื่นๆ ที่น่าสนใจ

```bash
# 🐛 Logs & Debugging
make logs        # ดู Log รวมทั้งหมด
make logs-api    # ดู Log เฉพาะ Backend API
make logs-db     # ดู Log เฉพาะ Database

# 🧪 Testing
make test        # รัน unit tests ทั้งหมด
make test-cover  # รันเทสและดูเปอร์เซ็นต์ Coverage
make test-html   # รันเทสและเปิดหน้าเว็บสรุป Coverage แบบสวยงาม
```

---

## 🌍 Deploy Online (Production)

ปัจจุบันเราแยก Frontend และ Backend ออกจากกันอย่างชัดเจน:

### 1. ฝั่ง Backend API (Railway / Render / อื่นๆ)
1. ตั้งค่า Database บน Cloud (เช่น Aiven MySQL) และนำ Connection String มาใส่ Environment Variables
2. Deploy โค้ดชุดนี้ขึ้น Railway โดยกำหนดให้ Railway ใช้ `Dockerfile.backend`
3. ตั้งค่า Env Variables ที่จำเป็น (ดูได้ใน `.env.example`)
   - `CORS_ALLOWED_ORIGIN=*` หรือกำหนดเป็น URL ของ Frontend 
4. รัน Migration ด้วยการตั้ง `RUN_MIGRATION=true` ใน Deploy ครั้งแรก

### 2. ฝั่ง Frontend (Vercel)
1. นำโฟลเดอร์ `frontend/` ไป Deploy ขึ้น Vercel
2. เนื่องจากมีไฟล์ `vercel.json` อยู่แล้ว ระบบจะทำการ Rewrite Routes ให้
3. **อย่าลืม!** ต้องตั้งค่า Environment Variable บน Vercel:
   `VITE_API_URL=https://<url-backend-ของคุณ>` เพื่อให้เว็บยิง API ไปหา Backend จริงได้สำเร็จ

---

## 🔒 ความปลอดภัย (Security Features)

- **Secret Ballot** — ไม่เชื่อม voter กับ candidate ที่เลือกในฐานข้อมูล
- **Citizen ID Hashing** — เก็บแค่ HMAC-SHA256
- **OTP Lockout** — กรอกผิด 5 ครั้งล็อกทันที
- **Row-level Locking** — ป้องกัน duplicate vote จาก concurrent requests
- **Rate Limiting** — 60 req/min/IP
- **Audit Log** — บันทึกทุก ballot submission

---

## License

MIT

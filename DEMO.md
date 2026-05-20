# VoteSpher — Demo Guide

## Prerequisites

- Backend running: `go run cmd/server/main.go`
- Frontend running: `cd frontend && npm run dev`
- `.env` ตั้ง `ENABLE_DEV_ENDPOINTS=true`
- เปิด browser: **http://localhost:3000**

---

## Test Accounts

| citizen_id      | voter_id | เขต   | email            | role  |
|-----------------|----------|-------|------------------|-------|
| 0000000000001   | 21       | เขต 1 | test01@test.com  | voter |
| 0000000000003   | 23       | เขต 2 | test03@test.com  | voter |
| 0000000000005   | 25       | เขต 3 | test05@test.com  | voter |
| 0000000000009   | 29       | เขต 1 | test09@test.com  | **admin** |

> OTP ทุก account ใช้ `111111` ใน dev mode

---

## Demo Scenarios

### 1. ดูผลโหวต (ไม่ต้อง login)

1. เปิด **http://localhost:3000**
2. กด **"ดูผลโหวต Realtime"** ด้านล่าง
3. เห็น dashboard แสดงผลแยกเขต / แยกพรรค

---

### 2. Voter — ลงคะแนนเสียง

1. กรอก `citizen_id`: **0000000000001**
2. เลือกช่องรับ OTP: **Email**
3. กด **ยืนยันตัวตน** → กด **ส่ง OTP ทาง Email**
4. กรอก OTP: **111111**
5. เห็นหน้า **ลงคะแนนเสียง** พร้อมรายชื่อผู้สมัครในเขต 1
6. เลือก candidate → กด **ลงคะแนน**
7. ระบบ redirect ไปหน้าผลโหวต Realtime อัตโนมัติ

> หลังโหวตแล้ว voter คนเดิม login ใหม่จะเข้าหน้าผลตรง ๆ (โหวตซ้ำไม่ได้)

---

### 3. Admin — ควบคุม Election State

1. กรอก `citizen_id`: **0000000000009**
2. OTP: **111111**
3. เข้า **Admin Dashboard** อัตโนมัติ
4. เห็น state ปัจจุบัน + ปุ่ม transition ที่ทำได้

| State     | ความหมาย         | ไปต่อได้           |
|-----------|------------------|--------------------|
| PREPARE   | เตรียมการ        | → OPEN             |
| OPEN      | เปิดโหวต         | → PAUSED / CLOSED  |
| PAUSED    | หยุดชั่วคราว     | → OPEN / CLOSED    |
| CLOSED    | ปิดโหวต          | → COUNTING         |
| COUNTING  | นับคะแนน (final) | —                  |

> Voter จะโหวตได้เฉพาะตอน state = **OPEN**

---

### 4. Realtime Results

- เปิดหน้าผลโหวตค้างไว้ขณะมี voter อื่น submit ballot
- ตัวเลขอัพเดต live ผ่าน polling (ไม่ต้อง refresh)

---

## Quick Bypass (ไม่ต้องผ่าน login flow)

```bash
# Voter
curl -X POST http://localhost:8080/dev/mock-token \
  -H "Content-Type: application/json" \
  -d '{"voter_id": 21, "area_id": 1, "role": "voter"}'

# Admin
curl -X POST http://localhost:8080/dev/mock-token \
  -H "Content-Type: application/json" \
  -d '{"voter_id": 29, "area_id": 1, "role": "admin"}'
```

ใช้ token ที่ได้ใส่ใน header: `Authorization: Bearer <token>`

---

## Key Features to Highlight

- **Secret Ballot** — ไม่มีทางเชื่อมระหว่าง voter กับ candidate ที่เลือก
- **OTP 2FA** — บัตรประชาชน + OTP ทาง Email/SMS
- **Duplicate vote prevention** — row-level lock, โหวตซ้ำไม่ได้แม้ส่ง concurrent requests
- **Election State Machine** — admin ควบคุม flow การเลือกตั้งได้
- **Realtime Dashboard** — ผลอัพเดต live ทุก voter

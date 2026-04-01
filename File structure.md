# File structure

---

```File structure
election-api/
├── cmd/
│   └── server/
│       └── main.go                      // ประกอบ dependency ทุกชิ้น, เริ่ม HTTP server + Graceful Shutdown
├── config/
│   └── config.go                        // อ่าน env vars → struct Config (DB, Port, JWT, Admin, Port)
├── internal/                            // โค้ด business หลัก — Go ป้องกันไม่ให้ package นอกโปรเจกต์ import
│   ├── domain/                          // structs กลาง + interfaces — ทุก layer ใช้ร่วมกัน
│   │   ├── voter.go                     // struct Voter, VoterInfo, AccessToken
│   │   ├── otp.go                       // struct OTP, OTPResult
│   │   ├── candidate.go                 // struct Candidate, Party
│   │   ├── ballot.go                    // struct Ballot, BallotResult, VoteStatus
│   │   ├── result.go                    // struct ElectionResult, ResultSummary, AreaResult
│   │   ├── system_config.go             // struct SystemConfig (is_voting_open, end_time)
│   │   ├── service_interfaces.go        // interface AuthService, OTPService, BallotService ฯลฯ — handler คุยผ่านนี้
│   │   └── repository_interfaces.go     // interface VoterRepo, OTPRepo, BallotRepo ฯลฯ — service คุยผ่านนี้
│   ├── handler/                         // layer 1: รับ HTTP, validate input, ส่งต่อ service
│   │   ├── auth_handler.go              // POST /voter/verify, /otp-request, /otp-confirm
│   │   ├── candidate_handler.go         // GET /candidates, GET /parties
│   │   ├── ballot_handler.go            // POST /ballot/submit, GET /ballot/status
│   │   ├── result_handler.go            // GET /results/realtime, GET /results/area/:id
│   │   └── admin_handler.go             // PATCH /election/config
│   ├── service/                         // layer 2: business logic ทั้งหมด
│   │   ├── auth_service.go              // hash citizen_id, ตรวจสิทธิ์, สร้าง JWT
│   │   ├── otp_service.go               // สร้าง/validate/invalidate OTP, เรียก SMS gateway
│   │   ├── candidate_service.go         // ดึงผู้สมัคร + join ข้อมูลพรรค
│   │   ├── ballot_service.go            // ตรวจ is_voted, บันทึก DB transaction, อัปเดต flag
│   │   ├── result_service.go            // aggregate คะแนน, ดึง/เขียน จาก MySQL
│   │   └── admin_service.go             // ตรวจ admin token, อัปเดต system config
│   ├── repository/                      // layer 3: คุยกับ DB โดยตรง
│   │   ├── voter_repo.go                // FindByHashAndArea(), UpdateVotedStatus()
│   │   ├── otp_repo.go                  // Save(), FindByVoterID(), Invalidate() — ใช้ MySQL ล้วน
│   │   ├── candidate_repo.go            // FindByArea() — query CANDIDATES join PARTIES
│   │   ├── ballot_repo.go               // SubmitWithTransaction() — INSERT VOTES + UPDATE VOTERS atomic
│   │   ├── result_repo.go               // aggregate COUNT GROUP BY, อ่าน/เขียน
│   │   └── config_repo.go               // GetConfig(), UpdateConfig()
│   ├── middleware/                      // ตรวจสอบก่อนเข้า handler
│   │   ├── jwt_middleware.go            // decode JWT, inject voter_id + area_id ลง Gin context
│   │   └── admin_middleware.go          // ตรวจ static admin token → 401 ถ้าไม่ตรง
│   └── router/                          // ลงทะเบียน routes ทั้งหมดในที่เดียว
│       └── router.go                    // gin.Group() แยกหมวด, ผูก middleware กับ group ที่ต้องการ
├── pkg/                                 // utility ไม่มี business logic — reuse ได้ทั่วโปรเจกต์
│   ├── hash/
│   │   └── sha256.go                    // Hash(input) string — ใช้ crypto/sha256 hash citizen_id
│   ├── jwt/
│   │   └── jwt.go                       // Encode(payload, secret), Decode(token, secret)
│   ├── otp/
│   │   └── generator.go                 // Generate() string — สุ่ม 6 หลักด้วย crypto/rand
│   ├── sms/
│   │   └── gateway.go                   // interface SMSGateway + MockSMSGateway สำหรับ test/dev
│   └── response/
│       └── response.go                  // Success(c, data), Error(c, code, msg) — JSON response มาตรฐาน
├── database/
│   ├── mysql.go
│   └── migrations/                      // SQL schema — รันตามลำดับด้วย golang-migrate
│       ├── 001_create_voters.sql         // ตาราง VOTERS (voter_id, citizen_id_hash, area_id, phone, is_voted)
│       ├── 002_create_candidates.sql     // ตาราง CANDIDATES (candidate_id, area_id, party_id, no, name)
│       ├── 003_create_parties.sql        // ตาราง PARTIES (party_id, party_name, logo_url)
│       ├── 004_create_votes.sql          // ตาราง VOTES (vote_id, voter_id, area_id, candidate_no, fingerprint)
│       └── 005_create_system_config.sql  // ตาราง SYSTEM_CONFIG (is_voting_open, end_time)
│       └── 006_create_otps.sql
├── .env.example                         // ตัวอย่าง env vars — ก็อปไปสร้าง .env จริง (ไม่ commit ค่าจริง)
├── .gitignore                           // ไม่ track .env, binary, vendor/, tmp/
├── go.mod                               // ชื่อ module + Go version + direct dependencies
├── go.sum                               // checksum ทุก dependency — ไม่แก้มือ
└── Makefile                             // make run, make build, make migrate, make test, make mock  ในส่วนของ database ถ้าใช้เป็น sql จะต้องเขียนยังไง
```

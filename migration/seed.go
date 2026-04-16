package migration

import (
	"log"
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

func uintPtr(v uint) *uint {
	return &v
}

func timePtr(timeStr string) *time.Time {
	if timeStr == "NULL" || timeStr == "" {
		return nil
	}
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		log.Printf("Time parse error: %v", err)
		return nil
	}
	return &t
}

func parseTime(timeStr string) time.Time {
	layout := "2006-01-02 15:04:05"
	t, _ := time.Parse(layout, timeStr)
	return t
}

func SeedData(db *gorm.DB) {
    var count int64
    db.Model(&models.Voter{}).Count(&count)
    if count > 0 {
        log.Println("✅ Data already seeded. Skipping...")
        return
    }

    log.Println("🧹 Clearing old data and resetting tables...")

    db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
    db.Exec("TRUNCATE TABLE votes;")
    db.Exec("TRUNCATE TABLE system_configs;")
    db.Exec("TRUNCATE TABLE admins;")
    db.Exec("TRUNCATE TABLE otps;")
    db.Exec("TRUNCATE TABLE voters;")
    db.Exec("TRUNCATE TABLE candidates;")
    db.Exec("TRUNCATE TABLE parties;")
    db.Exec("TRUNCATE TABLE areas;")
    db.Exec("SET FOREIGN_KEY_CHECKS = 1;")

    log.Println("🌱 Seeding Mock Data...")

    // 1. Areas
    areas := []models.Area{
        {AreaName: "กรุงเทพมหานคร เขต 1"},
        {AreaName: "กรุงเทพมหานคร เขต 2"},
        {AreaName: "เชียงใหม่ เขต 1"},
        {AreaName: "ขอนแก่น เขต 1"},
        {AreaName: "ชลบุรี เขต 1"},
    }
    if err := db.Create(&areas).Error; err != nil {
        log.Fatalf("❌ Failed to seed areas: %v", err)
    }

    // 2. Parties
    parties := []models.Party{
        {PartyNo: 1, PartyName: "พรรคประชาชน", LogoURL: "https://media.thairath.co.th/image/JRBsXLw1vXQ5sB7aeUwepRqs1a2hJoWnkygtep1V2yo0tUoPpoe9vA1.jpg"},
        {PartyNo: 2, PartyName: "พรรคเพื่อไทย", LogoURL: "https://yt3.googleusercontent.com/dtPvavXvlxB6kiY4LB1hUjrvVV6MlJuXV1yOjjBEkAEEM1jdbHanzuTitzS3L0HbUSlhbwaZOA=s900-c-k-c0x00ffffff-no-rj"},
        {PartyNo: 3, PartyName: "พรรคภูมิใจไทย", LogoURL: "https://www.infoquest.co.th/dxt-content/uploads/2025/06/20250618_-1024x576.png"},
        {PartyNo: 4, PartyName: "พรรคกล้าธรรม", LogoURL: "https://upload.wikimedia.org/wikipedia/th/thumb/c/cc/KlaThamParty_logo_%282025%29.jpg/250px-KlaThamParty_logo_%282025%29.jpg"},
    }
    if err := db.Create(&parties).Error; err != nil {
        log.Fatalf("❌ Failed to seed parties: %v", err)
    }

    // 3. Candidates (เปลี่ยน .AreaID / .PartyID เป็น .ID)
    candidates := []models.Candidate{
        {AreaID: areas[0].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายสมชาย ใจดี", Biography: "อดีตนักธุรกิจไฟแรง"},
        {AreaID: areas[0].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาวสมศรี มีทรัพย์", Biography: "นักเคลื่อนไหวเพื่อสังคม"},
        {AreaID: areas[0].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายวิชัย เก่งกล้า", Biography: "ทนายความอาสา"},
        {AreaID: areas[0].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นางมะลิ หอมหวน", Biography: "อดีตครูดีเด่น"},
        {AreaID: areas[1].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายอำนาจ มั่นคง", Biography: "วิศวกรซอฟต์แวร์"},
        {AreaID: areas[1].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาววิไล สวยงาม", Biography: "แพทย์หญิงจิตอาสา"},
        {AreaID: areas[1].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายธนาธร รักชาติ", Biography: "นักวิชาการอิสระ"},
        {AreaID: areas[1].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นางสุดา ยิ้มแย้ม", Biography: "ประธานชุมชน"},
        {AreaID: areas[2].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายยอดชาย ชัยชนะ", Biography: "เกษตรกรยุคใหม่"},
        {AreaID: areas[2].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาวดวงใจ ใสสะอาด", Biography: "นักสิ่งแวดล้อม"},
        {AreaID: areas[2].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายกฤษดา มานะ", Biography: "นักเขียนชื่อดัง"},
        {AreaID: areas[2].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นางจันทร์เพ็ญ เด่นดวง", Biography: "อดีตข้าราชการ"},
        {AreaID: areas[3].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายเอกชัย ใจสู้", Biography: "ผู้นำแรงงาน"},
        {AreaID: areas[3].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาวพรทิพย์ ริบรู้", Biography: "อาจารย์มหาวิทยาลัย"},
        {AreaID: areas[3].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายสมปอง ทองดี", Biography: "นักธุรกิจท้องถิ่น"},
        {AreaID: areas[3].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นางยุพา พาสุข", Biography: "พยาบาลวิชาชีพ"},
        {AreaID: areas[4].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายวรวุฒิ สุดยอด", Biography: "ตัวแทนคนรุ่นใหม่"},
        {AreaID: areas[4].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาวรัตนา นารี", Biography: "นักกีฬาเหรียญทอง"},
        {AreaID: areas[4].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายพิภพ จบจริง", Biography: "ทนายความสิทธิมนุษยชน"},
        {AreaID: areas[4].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นางสายใจ สายเสมอ", Biography: "แม่ค้าตลาดสด"},
    }
    if err := db.Create(&candidates).Error; err != nil {
        log.Fatalf("❌ Failed to seed candidates: %v", err)
    }

    // 4. Voters (เปลี่ยน .AreaID เป็น .ID)
    voters := []models.Voter{
        {CitizenIDHash: "342dcd58481f26e0109717a0930621f769707e9091da3697a2312c8973f1b130", AreaID: areas[0].ID, PhoneNumber: "0811111111", IsVoted: true, VotedAt: timePtr("2026-04-14 08:30:00")},
        {CitizenIDHash: "59194e4bf88deb95b5b4e6eb6b09463bdc1592aa13c943f470a148b98974e8ab", AreaID: areas[0].ID, PhoneNumber: "0812222222", IsVoted: true, VotedAt: timePtr("2026-04-14 09:15:00")},
        {CitizenIDHash: "02cde367f2cdd9210f7edb4ddf486032067f7067ac84891a22efa7d9b77de8af", AreaID: areas[0].ID, PhoneNumber: "0813333333", IsVoted: true, VotedAt: timePtr("2026-04-14 10:05:00")},
        {CitizenIDHash: "a69e7a3ee0ac644de72323c3932c528921fdf8319630470f753dbcdd09e7becb", AreaID: areas[0].ID, PhoneNumber: "0814444444", IsVoted: false, VotedAt: nil},
        {CitizenIDHash: "5aba4c653982bfe235c2213b754348b50e8f2d61be0010a6923f18032102dc15", AreaID: areas[1].ID, PhoneNumber: "0821111111", IsVoted: true, VotedAt: timePtr("2026-04-14 08:45:00")},
        {CitizenIDHash: "051b71582088500660bc221f9b2c3971820a798f524a69bade02a24a95e5417e", AreaID: areas[1].ID, PhoneNumber: "0822222222", IsVoted: true, VotedAt: timePtr("2026-04-14 11:20:00")},
        {CitizenIDHash: "0afe2c335af6c6c40d6e93b905031728617bc385e16bea3b81b26fb887f76ca1", AreaID: areas[1].ID, PhoneNumber: "0823333333", IsVoted: true, VotedAt: timePtr("2026-04-14 13:10:00")},
        {CitizenIDHash: "cd7ccccce8894be61b784d0daffd2f169f491d03822bd9da7b8bc62585eab14c", AreaID: areas[1].ID, PhoneNumber: "0824444444", IsVoted: true, VotedAt: timePtr("2026-04-14 14:00:00")},
        {CitizenIDHash: "bf11eedf5b0deb55cc1e8df88663559d47a5186c54f394137b8b0126ccbf8637", AreaID: areas[2].ID, PhoneNumber: "0831111111", IsVoted: true, VotedAt: timePtr("2026-04-14 09:00:00")},
        {CitizenIDHash: "cbe22b81529cdf737479c3c7671b217dfda91a263f0c39cda62c1961767fc9fc", AreaID: areas[2].ID, PhoneNumber: "0832222222", IsVoted: true, VotedAt: timePtr("2026-04-14 09:30:00")},
        {CitizenIDHash: "14193cf706990e9ff91242210b365e85faed4135e74c6729d6aa4ed01f94f410", AreaID: areas[2].ID, PhoneNumber: "0833333333", IsVoted: false, VotedAt: nil},
        {CitizenIDHash: "b653bbe2bef5e72cdc77e2ee34ce505f82ee9ab699136a5b35e3cea426261fbd", AreaID: areas[2].ID, PhoneNumber: "0834444444", IsVoted: true, VotedAt: timePtr("2026-04-14 10:45:00")},
        {CitizenIDHash: "5ffcfda5e9c4114eced76d646230c64080106ce534aa3e7d869ae516449dfb9a", AreaID: areas[3].ID, PhoneNumber: "0841111111", IsVoted: true, VotedAt: timePtr("2026-04-14 08:10:00")},
        {CitizenIDHash: "42f72dd8c2cee780500e514a11d9de6a76b0ede65741a5c3cb66f94a3229201d", AreaID: areas[3].ID, PhoneNumber: "0842222222", IsVoted: true, VotedAt: timePtr("2026-04-14 12:30:00")},
        {CitizenIDHash: "9a27504f65ece25808600e5f557bebaf2ab935d2597499602f9eb2d4111e7c39", AreaID: areas[3].ID, PhoneNumber: "0843333333", IsVoted: true, VotedAt: timePtr("2026-04-14 15:20:00")},
        {CitizenIDHash: "9efa6f667e8f08d85cc6106349ea7e93dea78028bc986dee10b8a4af098283a8", AreaID: areas[3].ID, PhoneNumber: "0844444444", IsVoted: true, VotedAt: timePtr("2026-04-14 16:05:00")},
        {CitizenIDHash: "179663769254b15efa21a26d0420b873db7b6b512b5d77422bd9c519b86ae395", AreaID: areas[4].ID, PhoneNumber: "0851111111", IsVoted: true, VotedAt: timePtr("2026-04-14 08:50:00")},
        {CitizenIDHash: "3555e05ef472d7071981b0d30d909a524ef97b17304a303574713bf2a9ffb2dd", AreaID: areas[4].ID, PhoneNumber: "0852222222", IsVoted: true, VotedAt: timePtr("2026-04-14 09:40:00")},
        {CitizenIDHash: "216675eb3dd157942ee9799eda1627eafa55a80fcf74f724e9d03412208bf6fe", AreaID: areas[4].ID, PhoneNumber: "0853333333", IsVoted: false, VotedAt: nil},
        {CitizenIDHash: "cee14863a6e22a02647c875d886a5d0adc6439e73b26bc8fe943cacdaab8677c", AreaID: areas[4].ID, PhoneNumber: "0854444444", IsVoted: true, VotedAt: timePtr("2026-04-14 11:15:00")},
    }
    if err := db.Create(&voters).Error; err != nil {
        log.Fatalf("❌ Failed to seed voters: %v", err)
    }

	// ==========================================
    // 4.5 OTPs
    // ==========================================
    otps := []models.OTP{
        {VoterID: voters[0].ID, OTPCode: "123456", RefCode: "ABCD", ExpiresAt: parseTime("2026-04-14 08:35:00"), IsUsed: true},
        {VoterID: voters[1].ID, OTPCode: "234567", RefCode: "EFGH", ExpiresAt: parseTime("2026-04-14 09:20:00"), IsUsed: true},
        {VoterID: voters[2].ID, OTPCode: "345678", RefCode: "IJKL", ExpiresAt: parseTime("2026-04-14 10:10:00"), IsUsed: true},
        {VoterID: voters[3].ID, OTPCode: "456789", RefCode: "MNOP", ExpiresAt: parseTime("2026-04-14 11:00:00"), IsUsed: false},
        {VoterID: voters[4].ID, OTPCode: "567890", RefCode: "QRST", ExpiresAt: parseTime("2026-04-14 08:50:00"), IsUsed: true},
        {VoterID: voters[5].ID, OTPCode: "678901", RefCode: "UVWX", ExpiresAt: parseTime("2026-04-14 11:25:00"), IsUsed: true},
        {VoterID: voters[6].ID, OTPCode: "789012", RefCode: "YZAB", ExpiresAt: parseTime("2026-04-14 13:15:00"), IsUsed: true},
        {VoterID: voters[7].ID, OTPCode: "890123", RefCode: "CDEF", ExpiresAt: parseTime("2026-04-14 14:05:00"), IsUsed: true},
        {VoterID: voters[8].ID, OTPCode: "901234", RefCode: "GHIJ", ExpiresAt: parseTime("2026-04-14 09:05:00"), IsUsed: true},
        {VoterID: voters[9].ID, OTPCode: "012345", RefCode: "KLMN", ExpiresAt: parseTime("2026-04-14 09:35:00"), IsUsed: true},
        {VoterID: voters[10].ID, OTPCode: "112233", RefCode: "OPQR", ExpiresAt: parseTime("2026-04-14 10:00:00"), IsUsed: false},
        {VoterID: voters[11].ID, OTPCode: "223344", RefCode: "STUV", ExpiresAt: parseTime("2026-04-14 10:50:00"), IsUsed: true},
        {VoterID: voters[12].ID, OTPCode: "334455", RefCode: "WXYZ", ExpiresAt: parseTime("2026-04-14 08:15:00"), IsUsed: true},
        {VoterID: voters[13].ID, OTPCode: "445566", RefCode: "ABCD", ExpiresAt: parseTime("2026-04-14 12:35:00"), IsUsed: true},
        {VoterID: voters[14].ID, OTPCode: "556677", RefCode: "EFGH", ExpiresAt: parseTime("2026-04-14 15:25:00"), IsUsed: true},
        {VoterID: voters[15].ID, OTPCode: "667788", RefCode: "IJKL", ExpiresAt: parseTime("2026-04-14 16:10:00"), IsUsed: true},
        {VoterID: voters[16].ID, OTPCode: "778899", RefCode: "MNOP", ExpiresAt: parseTime("2026-04-14 08:55:00"), IsUsed: true},
        {VoterID: voters[17].ID, OTPCode: "889900", RefCode: "QRST", ExpiresAt: parseTime("2026-04-14 09:45:00"), IsUsed: true},
        {VoterID: voters[18].ID, OTPCode: "990011", RefCode: "UVWX", ExpiresAt: parseTime("2026-04-14 10:30:00"), IsUsed: false},
        {VoterID: voters[19].ID, OTPCode: "001122", RefCode: "YZAB", ExpiresAt: parseTime("2026-04-14 11:20:00"), IsUsed: true}, // ปรับ OTP เป็น 6 หลักให้สมบูรณ์
    }
    if err := db.Create(&otps).Error; err != nil {
        log.Fatalf("❌ Failed to seed otps: %v", err)
    }

    // 5. Admins (เปลี่ยน .VoterID เป็น .ID)
    admins := []models.Admin{
        {VoterID: voters[0].ID},
        {VoterID: voters[1].ID},
    }
    if err := db.Create(&admins).Error; err != nil {
        log.Fatalf("❌ Failed to seed admins: %v", err)
    }

    // 6. SystemConfigs (เปลี่ยน .AdminID เป็น .ID)
    configs := []models.SystemConfig{
        {AdminID: admins[0].ID, Status: "OPEN", StartTime: parseTime("2026-04-14 08:00:00"), EndTime: parseTime("2026-04-28 17:00:00"), UpdatedAt: parseTime("2026-04-13 20:00:00"), IsActive: true},
    }
    if err := db.Create(&configs).Error; err != nil {
        log.Fatalf("❌ Failed to seed system_configs: %v", err)
    }

    // 7. Votes (เปลี่ยน .AreaID, .CandidateID, .PartyID เป็น .ID)
    votes := []models.Vote{
        {AreaID: areas[0].ID, CandidateID: &candidates[0].ID, PartyID: &parties[0].ID, CreatedAt: parseTime("2026-04-14 08:30:05")},
        {AreaID: areas[0].ID, CandidateID: &candidates[1].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 09:15:12")},
        {AreaID: areas[0].ID, CandidateID: &candidates[2].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 10:05:45")},
        {AreaID: areas[0].ID, CandidateID: &candidates[3].ID, PartyID: &parties[3].ID, CreatedAt: parseTime("2026-04-14 11:45:00")},
        {AreaID: areas[1].ID, CandidateID: nil, PartyID: &parties[0].ID, CreatedAt: parseTime("2026-04-14 11:20:10")},
        {AreaID: areas[1].ID, CandidateID: &candidates[6].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 13:10:05")},
        {AreaID: areas[1].ID, CandidateID: &candidates[7].ID, PartyID: &parties[3].ID, CreatedAt: parseTime("2026-04-14 14:00:55")},
        {AreaID: areas[2].ID, CandidateID: &candidates[8].ID, PartyID: &parties[0].ID, CreatedAt: parseTime("2026-04-14 09:00:23")},
        {AreaID: areas[2].ID, CandidateID: &candidates[9].ID, PartyID: nil, CreatedAt: parseTime("2026-04-14 09:30:44")},
        {AreaID: areas[2].ID, CandidateID: &candidates[10].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 10:45:11")},
        {AreaID: areas[3].ID, CandidateID: &candidates[13].ID, PartyID: &parties[1].ID, CreatedAt: parseTime("2026-04-14 12:30:33")},
        {AreaID: areas[3].ID, CandidateID: &candidates[14].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 15:20:19")},
        {AreaID: areas[3].ID, CandidateID: nil, PartyID: nil, CreatedAt: parseTime("2026-04-14 16:05:59")},
        {AreaID: areas[4].ID, CandidateID: &candidates[16].ID, PartyID: &parties[0].ID, CreatedAt: parseTime("2026-04-14 08:50:41")},
        {AreaID: areas[4].ID, CandidateID: &candidates[17].ID, PartyID: &parties[1].ID, CreatedAt: parseTime("2026-04-14 09:40:22")},
        {AreaID: areas[4].ID, CandidateID: &candidates[19].ID, PartyID: &parties[3].ID, CreatedAt: parseTime("2026-04-14 11:15:08")},
        {AreaID: areas[4].ID, CandidateID: &candidates[18].ID, PartyID: &parties[2].ID, CreatedAt: parseTime("2026-04-14 13:20:00")},
    }
    if err := db.Create(&votes).Error; err != nil {
        log.Fatalf("❌ Failed to seed votes: %v", err)
    }

    log.Println("✅ Database successfully seeded with all mock data!")
}
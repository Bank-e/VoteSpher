package migration

import (
	"log"
	"os"
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
	force := os.Getenv("FORCE_SEED") == "true"
	if !force {
		var count int64
		db.Model(&models.Voter{}).Count(&count)
		if count > 0 {
			log.Println("✅ Data already seeded. Skipping... (set FORCE_SEED=true to re-seed)")
			return
		}
	}

	log.Println("🧹 Clearing old data — drop + recreate affected tables...")
	db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	// Drop tables that may have schema mismatches, then let AutoMigrate recreate them
	db.Exec("DROP TABLE IF EXISTS votes;")
	db.Exec("DROP TABLE IF EXISTS audit_logs;")
	db.Exec("DROP TABLE IF EXISTS system_configs;")
	db.Exec("DROP TABLE IF EXISTS admins;")
	db.Exec("DROP TABLE IF EXISTS otps;")
	db.Exec("DROP TABLE IF EXISTS voters;")
	db.Exec("DROP TABLE IF EXISTS candidates;")
	db.Exec("DROP TABLE IF EXISTS parties;")
	db.Exec("DROP TABLE IF EXISTS areas;")
	db.Exec("DROP TABLE IF EXISTS provinces;")
	db.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	// Recreate with current schema
	if err := db.AutoMigrate(
		&models.Province{}, &models.Area{}, &models.Party{}, &models.Voter{},
		&models.Candidate{}, &models.OTP{}, &models.Admin{},
		&models.SystemConfig{}, &models.Vote{}, &models.AuditLog{},
	); err != nil {
		log.Fatalf("re-migrate failed: %v", err)
	}
	log.Println("✅ Tables recreated")

	log.Println("🌱 Seeding Base Mock Data...")

	// 1. Provinces
	provinces := []models.Province{{ProvinceName: "กรุงเทพมหานคร"}}
	if err := db.Create(&provinces).Error; err != nil {
		log.Fatalf("❌ Failed to seed provinces: %v", err)
	}

	// 2. Areas (33 เขตเลือกตั้ง กทม.)
	areas := []models.Area{
		{AreaName: "เขต 1", ProvinceID: provinces[0].ID}, {AreaName: "เขต 2", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 3", ProvinceID: provinces[0].ID}, {AreaName: "เขต 4", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 5", ProvinceID: provinces[0].ID}, {AreaName: "เขต 6", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 7", ProvinceID: provinces[0].ID}, {AreaName: "เขต 8", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 9", ProvinceID: provinces[0].ID}, {AreaName: "เขต 10", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 11", ProvinceID: provinces[0].ID}, {AreaName: "เขต 12", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 13", ProvinceID: provinces[0].ID}, {AreaName: "เขต 14", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 15", ProvinceID: provinces[0].ID}, {AreaName: "เขต 16", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 17", ProvinceID: provinces[0].ID}, {AreaName: "เขต 18", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 19", ProvinceID: provinces[0].ID}, {AreaName: "เขต 20", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 21", ProvinceID: provinces[0].ID}, {AreaName: "เขต 22", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 23", ProvinceID: provinces[0].ID}, {AreaName: "เขต 24", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 25", ProvinceID: provinces[0].ID}, {AreaName: "เขต 26", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 27", ProvinceID: provinces[0].ID}, {AreaName: "เขต 28", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 29", ProvinceID: provinces[0].ID}, {AreaName: "เขต 30", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 31", ProvinceID: provinces[0].ID}, {AreaName: "เขต 32", ProvinceID: provinces[0].ID},
		{AreaName: "เขต 33", ProvinceID: provinces[0].ID},
	}
	if err := db.Create(&areas).Error; err != nil {
		log.Fatalf("❌ Failed to seed areas: %v", err)
	}

	// 3. Parties
	parties := []models.Party{
		{PartyNo: 1, PartyName: "พรรคประชาชน", LogoURL: "/images/parties/party1_prachacha.jpg"},
		{PartyNo: 2, PartyName: "พรรคเพื่อไทย", LogoURL: "/images/parties/party2_phuathai.jpg"},
		{PartyNo: 3, PartyName: "พรรคภูมิใจไทย", LogoURL: "/images/parties/party3_bhumjaithai.jpg"},
		{PartyNo: 4, PartyName: "พรรคกล้าธรรม", LogoURL: "/images/parties/party4_klatharm.jpg"},
	}
	if err := db.Create(&parties).Error; err != nil {
		log.Fatalf("❌ Failed to seed parties: %v", err)
	}

	// 4. Candidates (จาก seed_candidates.go)
	candidates := SeedCandidates(db, areas, parties)

	// 5. Voters (เก็บ CitizenIDHash เดิมจาก main branch เพื่อให้ test accounts ยังใช้ได้)
	voters := []models.Voter{
		{CitizenIDHash: "342dcd58481f26e0109717a0930621f769707e9091da3697a2312c8973f1b130", AreaID: areas[0].ID, Email: "voter01@test.com", PhoneNumber: "0811111111", IsVoted: false},
		{CitizenIDHash: "59194e4bf88deb95b5b4e6eb6b09463bdc1592aa13c943f470a148b98974e8ab", AreaID: areas[0].ID, Email: "voter02@test.com", PhoneNumber: "0812222222", IsVoted: false},
		{CitizenIDHash: "02cde367f2cdd9210f7edb4ddf486032067f7067ac84891a22efa7d9b77de8af", AreaID: areas[0].ID, Email: "voter03@test.com", PhoneNumber: "0813333333", IsVoted: false},
		// voter_id 4: hash ของ citizen 1100100000001 → test account หลัก (email จริง)
		{CitizenIDHash: "a69e7a3ee0ac644de72323c3932c528921fdf8319630470f753dbcdd09e7becb", AreaID: areas[0].ID, Email: "piyachat.sal@dome.tu.ac.th", PhoneNumber: "0929400592", IsVoted: false},
		{CitizenIDHash: "5aba4c653982bfe235c2213b754348b50e8f2d61be0010a6923f18032102dc15", AreaID: areas[1].ID, Email: "voter05@test.com", PhoneNumber: "0821111111", IsVoted: false},
		{CitizenIDHash: "051b71582088500660bc221f9b2c3971820a798f524a69bade02a24a95e5417e", AreaID: areas[1].ID, Email: "voter06@test.com", PhoneNumber: "0822222222", IsVoted: false},
		{CitizenIDHash: "0afe2c335af6c6c40d6e93b905031728617bc385e16bea3b81b26fb887f76ca1", AreaID: areas[1].ID, Email: "voter07@test.com", PhoneNumber: "0823333333", IsVoted: false},
		{CitizenIDHash: "cd7ccccce8894be61b784d0daffd2f169f491d03822bd9da7b8bc62585eab14c", AreaID: areas[1].ID, Email: "voter08@test.com", PhoneNumber: "0824444444", IsVoted: false},
		{CitizenIDHash: "bf11eedf5b0deb55cc1e8df88663559d47a5186c54f394137b8b0126ccbf8637", AreaID: areas[2].ID, Email: "voter09@test.com", PhoneNumber: "0831111111", IsVoted: false},
		{CitizenIDHash: "cbe22b81529cdf737479c3c7671b217dfda91a263f0c39cda62c1961767fc9fc", AreaID: areas[2].ID, Email: "voter10@test.com", PhoneNumber: "0832222222", IsVoted: false},
		{CitizenIDHash: "14193cf706990e9ff91242210b365e85faed4135e74c6729d6aa4ed01f94f410", AreaID: areas[2].ID, Email: "voter11@test.com", PhoneNumber: "0833333333", IsVoted: false},
		{CitizenIDHash: "b653bbe2bef5e72cdc77e2ee34ce505f82ee9ab699136a5b35e3cea426261fbd", AreaID: areas[2].ID, Email: "voter12@test.com", PhoneNumber: "0834444444", IsVoted: false},
		{CitizenIDHash: "5ffcfda5e9c4114eced76d646230c64080106ce534aa3e7d869ae516449dfb9a", AreaID: areas[3].ID, Email: "voter13@test.com", PhoneNumber: "0841111111", IsVoted: false},
		{CitizenIDHash: "42f72dd8c2cee780500e514a11d9de6a76b0ede65741a5c3cb66f94a3229201d", AreaID: areas[3].ID, Email: "voter14@test.com", PhoneNumber: "0842222222", IsVoted: false},
		{CitizenIDHash: "9a27504f65ece25808600e5f557bebaf2ab935d2597499602f9eb2d4111e7c39", AreaID: areas[3].ID, Email: "voter15@test.com", PhoneNumber: "0843333333", IsVoted: false},
		{CitizenIDHash: "9efa6f667e8f08d85cc6106349ea7e93dea78028bc986dee10b8a4af098283a8", AreaID: areas[3].ID, Email: "voter16@test.com", PhoneNumber: "0844444444", IsVoted: false},
		{CitizenIDHash: "179663769254b15efa21a26d0420b873db7b6b512b5d77422bd9c519b86ae395", AreaID: areas[4].ID, Email: "voter17@test.com", PhoneNumber: "0851111111", IsVoted: false},
		{CitizenIDHash: "3555e05ef472d7071981b0d30d909a524ef97b17304a303574713bf2a9ffb2dd", AreaID: areas[4].ID, Email: "voter18@test.com", PhoneNumber: "0852222222", IsVoted: false},
		{CitizenIDHash: "216675eb3dd157942ee9799eda1627eafa55a80fcf74f724e9d03412208bf6fe", AreaID: areas[4].ID, Email: "voter19@test.com", PhoneNumber: "0853333333", IsVoted: false},
		{CitizenIDHash: "cee14863a6e22a02647c875d886a5d0adc6439e73b26bc8fe943cacdaab8677c", AreaID: areas[4].ID, Email: "voter20@test.com", PhoneNumber: "0854444444", IsVoted: false},
	}
	if err := db.Create(&voters).Error; err != nil {
		log.Fatalf("❌ Failed to seed voters: %v", err)
	}

	// 6. Admins (voter[3] เป็น admin — test account หลัก)
	admins := []models.Admin{
		{VoterID: voters[3].ID},
	}
	if err := db.Create(&admins).Error; err != nil {
		log.Fatalf("❌ Failed to seed admins: %v", err)
	}

	// 7. SystemConfigs
	configs := []models.SystemConfig{
		{AdminID: admins[0].ID, Status: "OPEN", StartTime: parseTime("2026-04-14 08:00:00"), EndTime: parseTime("2026-12-31 17:00:00"), UpdatedAt: parseTime("2026-04-13 20:00:00"), IsActive: true},
	}
	if err := db.Create(&configs).Error; err != nil {
		log.Fatalf("❌ Failed to seed system_configs: %v", err)
	}

	// 8. Votes 20,000 รายการ (จาก seed_votes.go)
	SeedVotes(db, areas, parties, candidates)

	log.Println("🎉 Database successfully seeded with ALL modular mock data!")
}

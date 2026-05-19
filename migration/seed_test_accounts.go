package migration

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"votespher/internal/models"

	"gorm.io/gorm"
)

func hashCitizenIDForSeed(citizenID string) string {
	key := []byte(os.Getenv("HASH_SECRET_KEY"))
	h := hmac.New(sha256.New, key)
	h.Write([]byte(citizenID))
	return hex.EncodeToString(h.Sum(nil))
}

// SeedTestAccounts adds 10 fresh test voters (is_voted=false) and fixes the admins table.
// Safe to run on a live DB — does NOT drop any tables or touch existing votes.
func SeedTestAccounts(db *gorm.DB) {
	log.Println("🧪 Adding 10 test accounts...")

	// Lookup area IDs by name
	areaIDForName := func(name string) uint {
		var area models.Area
		db.Where("area_name = ?", name).First(&area)
		if area.ID == 0 {
			log.Fatalf("❌ Area not found: %s", name)
		}
		return area.ID
	}

	area1 := areaIDForName("เขต 1")
	area2 := areaIDForName("เขต 2")
	area3 := areaIDForName("เขต 3")
	area4 := areaIDForName("เขต 4")
	area5 := areaIDForName("เขต 5")

	type testAccount struct {
		citizenID string
		email     string
		phone     string
		areaID    uint
	}

	accounts := []testAccount{
		{"0000000000001", "test01@test.com", "0900000001", area1},
		{"0000000000002", "test02@test.com", "0900000002", area1},
		{"0000000000003", "test03@test.com", "0900000003", area2},
		{"0000000000004", "test04@test.com", "0900000004", area2},
		{"0000000000005", "test05@test.com", "0900000005", area3},
		{"0000000000006", "test06@test.com", "0900000006", area3},
		{"0000000000007", "test07@test.com", "0900000007", area4},
		{"0000000000008", "test08@test.com", "0900000008", area4},
		{"0000000000009", "test09@test.com", "0900000009", area1}, // admin
		{"0000000000010", "test10@test.com", "0900000010", area5},
	}

	var newVoters []models.Voter
	for _, acc := range accounts {
		hash := hashCitizenIDForSeed(acc.citizenID)

		// Skip if citizen ID hash already exists
		var count int64
		db.Model(&models.Voter{}).Where("citizen_id_hash = ?", hash).Count(&count)
		if count > 0 {
			log.Printf("⚠️  citizen_id %s already exists — skipping", acc.citizenID)
			continue
		}

		newVoters = append(newVoters, models.Voter{
			CitizenIDHash: hash,
			AreaID:        acc.areaID,
			Email:         acc.email,
			PhoneNumber:   acc.phone,
			IsVoted:       false,
		})
	}

	if len(newVoters) > 0 {
		if err := db.Create(&newVoters).Error; err != nil {
			log.Fatalf("❌ Failed to insert test voters: %v", err)
		}
		log.Printf("✅ Inserted %d new test voters", len(newVoters))
	}

	// Find the admin voter (citizen 0000000000009)
	adminHash := hashCitizenIDForSeed("0000000000009")
	var adminVoter models.Voter
	if err := db.Where("citizen_id_hash = ?", adminHash).First(&adminVoter).Error; err != nil {
		log.Fatalf("❌ Admin voter not found: %v", err)
	}

	// Fix admins table
	log.Println("🔧 Fixing admins table...")

	// 1. Ensure new admin record exists
	var adminCount int64
	db.Model(&models.Admin{}).Where("voter_id = ?", adminVoter.ID).Count(&adminCount)
	if adminCount == 0 {
		if err := db.Create(&models.Admin{VoterID: adminVoter.ID}).Error; err != nil {
			log.Fatalf("❌ Failed to insert admin: %v", err)
		}
	}

	// 2. Get new admin's admin_id
	var newAdminRecord models.Admin
	db.Where("voter_id = ?", adminVoter.ID).First(&newAdminRecord)

	// 3. Reassign all system_configs to the new admin (releases FK from old records)
	db.Exec("UPDATE system_configs SET admin_id = ?", newAdminRecord.ID)

	// 4. Now safe to delete all other admin records
	result := db.Exec("DELETE FROM admins WHERE voter_id != ?", adminVoter.ID)
	log.Printf("✅ Admin set to voter_id=%d (citizen 0000000000009), removed %d old admin entries", adminVoter.ID, result.RowsAffected)

	log.Println("🎉 Done! Test accounts ready.")
	log.Println("")
	log.Println("  citizen_id       role   area   email")
	log.Println("  0000000000001    voter  เขต 1  test01@test.com")
	log.Println("  0000000000002    voter  เขต 1  test02@test.com")
	log.Println("  0000000000003    voter  เขต 2  test03@test.com")
	log.Println("  0000000000004    voter  เขต 2  test04@test.com")
	log.Println("  0000000000005    voter  เขต 3  test05@test.com")
	log.Println("  0000000000006    voter  เขต 3  test06@test.com")
	log.Println("  0000000000007    voter  เขต 4  test07@test.com")
	log.Println("  0000000000008    voter  เขต 4  test08@test.com")
	log.Println("  0000000000009    ADMIN  เขต 1  test09@test.com")
	log.Println("  0000000000010    voter  เขต 5  test10@test.com")
	log.Println("")
	log.Println("  OTP = 111111 (dev mode)")
}

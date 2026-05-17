package migration

import (
	"log"
	"math/rand"
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// Helper สำหรับ parse time ในไฟล์นี้ถ้าจำเป็น (ถ้าใช้จาก seed.go ได้ก็ไม่เป็นไร แต่เผื่อไว้)
func parseTimeLocal(timeStr string) time.Time {
	layout := "2006-01-02 15:04:05"
	t, _ := time.Parse(layout, timeStr)
	return t
}

func SeedVotes(db *gorm.DB, areas []models.Area, parties []models.Party, candidates []models.Candidate) {
	log.Println("🗳️ Generating 20,000 Mock Votes...")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalVotesTarget := 20000 // จำนวนโหวด
	numAreas := len(areas)
	avgVotes := totalVotesTarget / numAreas   // ~606
	minVotes := int(float64(avgVotes) * 0.90) // ขีดจำกัดล่าง -10%
	maxVotes := int(float64(avgVotes) * 1.10) // ขีดจำกัดบน +10%

	// 1. อัลกอริทึมคำนวณโควต้าจำนวนผู้มาใช้สิทธิในแต่ละเขต (ให้ได้ผลรวม 20,000 เป๊ะ)
	votesPerArea := make([]int, numAreas)
	totalAssigned := 0

	for i := 0; i < numAreas; i++ {
		votesPerArea[i] = minVotes
		totalAssigned += minVotes
	}

	for totalAssigned < totalVotesTarget {
		idx := r.Intn(numAreas)
		if votesPerArea[idx] < maxVotes {
			votesPerArea[idx]++
			totalAssigned++
		}
	}

	// 2. เตรียมตัวแปรก่อนการสุ่ม
	var votes []models.Vote
	electionStart := parseTimeLocal("2026-04-14 08:00:00")
	electionDuration := 9 * time.Hour // สุ่มเวลาตั้งแต่ 08:00 ถึง 17:00

	// 3. เริ่มสร้างข้อมูล Vote ตามโควต้าของแต่ละเขต
	for i, area := range areas {
		numVotesForThisArea := votesPerArea[i]

		// ค้นหาผู้สมัครทั้งหมดที่อยู่ในเขตนี้
		var areaCandidates []models.Candidate
		for _, c := range candidates {
			if c.AreaID == area.ID {
				areaCandidates = append(areaCandidates, c)
			}
		}

		// คำนวณน้ำหนักการโหวตของผู้สมัครในเขตนี้
		totalWeight := 0
		candidateWeights := make([]int, len(areaCandidates))
		for j, c := range areaCandidates {
			weight := 5 // น้ำหนักพื้นฐาน
			if c.PartyID == parties[0].ID {
				weight = 40
			} // ประชาชน 40%
			if c.PartyID == parties[1].ID {
				weight = 30
			} // เพื่อไทย 30%
			if c.PartyID == parties[2].ID {
				weight = 15
			} // ภูมิใจไทย 15%
			if c.PartyID == parties[3].ID {
				weight = 10
			} // กล้าธรรม 10%

			candidateWeights[j] = weight
			totalWeight += weight
		}

		// เพิ่มน้ำหนักสำหรับ "ไม่ประสงค์ลงคะแนน/บัตรเสีย" ประมาณ 5% ของกระดาน
		noVoteWeight := (totalWeight * 5) / 95
		if noVoteWeight == 0 {
			noVoteWeight = 1
		}
		totalWeight += noVoteWeight

		// =========================================
		// โอกาสของแต่ละ Pattern (รวม = 100)
		// tt = เลือกทั้งคนและพรรค     ~95%
		// th = เลือกคนอย่างเดียว      ~2%
		// ht = เลือกพรรคอย่างเดียว   ~2%
		// hh = บัตรเสีย/ไม่ประสงค์    ~1%
		// =========================================
		const patternHH = 1
		const patternHT = 2
		const patternTH = 2
		// patternTT = เหลือ 95%

		// สุ่มโหวต
		for v := 0; v < numVotesForThisArea; v++ {
			randomOffset := time.Duration(r.Int63n(int64(electionDuration)))
			voteTime := electionStart.Add(randomOffset)

			var cID *uint
			var pID *uint

			// ขั้นที่ 1: สุ่ม pattern ของบัตร
			patternRoll := r.Intn(100)

			if patternRoll < patternHH {
				// hh — บัตรเสีย: cID=nil, pID=nil (ไม่ต้องทำอะไร)

			} else {
				// ขั้นที่ 2: เลือก candidate โดยใช้ weighted random (เหมือนเดิม)
				var selectedCandidate *models.Candidate
				if len(areaCandidates) > 0 {
					roll := r.Intn(totalWeight - noVoteWeight) // ไม่รวม noVote slot
					currentWeight := 0
					for j, c := range areaCandidates {
						currentWeight += candidateWeights[j]
						if roll < currentWeight {
							tmp := c
							selectedCandidate = &tmp
							break
						}
					}
				}

				switch {
				case patternRoll < patternHH+patternHT:
					// ht — เลือกพรรคอย่างเดียว: cID=nil, pID=พรรคของผู้สมัครที่สุ่มได้
					if selectedCandidate != nil {
						pid := selectedCandidate.PartyID
						pID = &pid
					}

				case patternRoll < patternHH+patternHT+patternTH:
					// th — เลือกคนอย่างเดียว: cID=ผู้สมัคร, pID=nil
					if selectedCandidate != nil {
						cid := selectedCandidate.ID
						cID = &cid
					}

				default:
					// tt — เลือกทั้งคนและพรรค (ปกติ)
					if selectedCandidate != nil {
						cid := selectedCandidate.ID
						pid := selectedCandidate.PartyID
						cID = &cid
						pID = &pid
					}
				}
			}

			votes = append(votes, models.Vote{
				AreaID:      area.ID,
				CandidateID: cID,
				PartyID:     pID,
				CreatedAt:   voteTime,
			})
		}
	}

	log.Printf("⏳ Inserting %d votes into the database (Batch size: 1000)...", len(votes))
	if err := db.CreateInBatches(&votes, 1000).Error; err != nil {
		log.Fatalf("❌ Failed to seed massive votes: %v", err)
	}

	log.Printf("✅ %d Mock Votes seeded successfully!", totalVotesTarget)
}
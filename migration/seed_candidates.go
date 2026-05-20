package migration

import (
	"log"
	"votespher/internal/models"

	"gorm.io/gorm"
)

func SeedCandidates(db *gorm.DB, areas []models.Area, parties []models.Party) []models.Candidate {
	log.Println("🌱 Seeding Candidates...")

	candidates := []models.Candidate{
		// ==========================================
		// พรรคประชาชน (PartyID: parties[0].ID)
		// ==========================================
		{AreaID: areas[0].ID, PartyID: parties[0].ID, CandidateNo: 5, FullName: "นายปารเมศ วิทยารักษ์สรรค์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[1].ID, PartyID: parties[0].ID, CandidateNo: 2, FullName: "นายเสกสิทธิ์ แย้มสงวนศักดิ์", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนางสาวธิษะณา"},
		{AreaID: areas[2].ID, PartyID: parties[0].ID, CandidateNo: 9, FullName: "นายจรยุทธ จตุรพรประสิทธิ์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[3].ID, PartyID: parties[0].ID, CandidateNo: 2, FullName: "น.ส.ภัณฑิล น่วมเจิม", Biography: "อดีต สส. ในพื้นที่เศรษฐกิจหลัก"},
		{AreaID: areas[4].ID, PartyID: parties[0].ID, CandidateNo: 6, FullName: "นายปิติกรณ์ บรรณเภสัช", Biography: "ผู้สมัครหน้าใหม่ อดีตทีมงานเบื้องหลัง"},
		{AreaID: areas[5].ID, PartyID: parties[0].ID, CandidateNo: 7, FullName: "นายกันตภณ ดวงอัมพร", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[6].ID, PartyID: parties[0].ID, CandidateNo: 8, FullName: "น.ส.ภัสริน รามวงศ์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[7].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายไชยพล สท้อนดี", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[8].ID, PartyID: parties[0].ID, CandidateNo: 5, FullName: "นายศุภณัฐ มีนชัยนันท์", Biography: "อดีต สส. ฐานเสียงแข็งแกร่ง"},
		{AreaID: areas[9].ID, PartyID: parties[0].ID, CandidateNo: 6, FullName: "นายเอกราช อุดมอำนวย", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[10].ID, PartyID: parties[0].ID, CandidateNo: 15, FullName: "นายศศินันท์ ธรรมนิฐินันท์", Biography: "อดีต สส. แข่งขันเดือดกับกลุ่มสายไหมต้องรอด"},
		{AreaID: areas[11].ID, PartyID: parties[0].ID, CandidateNo: 15, FullName: "นายภูริวรรธก์ ใจสำราญ", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[12].ID, PartyID: parties[0].ID, CandidateNo: 7, FullName: "นายธนเดช เพ็งสุข", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[13].ID, PartyID: parties[0].ID, CandidateNo: 14, FullName: "นายก่อเกียรติ ก่อสูงศักดิ์", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนางสาวสิริลภัส"},
		{AreaID: areas[14].ID, PartyID: parties[0].ID, CandidateNo: 8, FullName: "นายวิทวัส ติชะวาณิชย์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[15].ID, PartyID: parties[0].ID, CandidateNo: 10, FullName: "พิมพ์กาญจน์ กีรติวิราปกรณ์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[16].ID, PartyID: parties[0].ID, CandidateNo: 10, FullName: "นายวีรวุธ รักเที่ยง", Biography: "อดีต สส. พื้นที่กึ่งชนบท"},
		{AreaID: areas[17].ID, PartyID: parties[0].ID, CandidateNo: 10, FullName: "นายธีรัจชัย พันธุมาศ", Biography: "อดีต สส. รุ่นใหญ่ของพรรค"},
		{AreaID: areas[18].ID, PartyID: parties[0].ID, CandidateNo: 6, FullName: "นายกันตพงษ์ ประยูรศักดิ์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[19].ID, PartyID: parties[0].ID, CandidateNo: 7, FullName: "นายชุมพล หลักคำ", Biography: "ผู้ท้าชิงที่ถูกส่งมาแก้มือกับพรรคเพื่อไทย"},
		{AreaID: areas[20].ID, PartyID: parties[0].ID, CandidateNo: 8, FullName: "นายณัฐพงศ์ เปรมพูลสวัสดิ์", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[21].ID, PartyID: parties[0].ID, CandidateNo: 6, FullName: "นายสุภกร ตันติไพบูลย์ธนะ", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[22].ID, PartyID: parties[0].ID, CandidateNo: 15, FullName: "นายชลธาร ทรัพย์ไพบูลย์เลิศ", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนายปิยรัฐ (โตโต้)"},
		{AreaID: areas[23].ID, PartyID: parties[0].ID, CandidateNo: 6, FullName: "นายณพัฏน์ จิตตภินันท์กัณตา", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนายเท่าพิภพ"},
		{AreaID: areas[24].ID, PartyID: parties[0].ID, CandidateNo: 7, FullName: "แอนศิริ วลัยกนก", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[25].ID, PartyID: parties[0].ID, CandidateNo: 1, FullName: "นายพงษ์สรณัฐ ทองลี", Biography: "ผู้สมัครหน้าใหม่"},
		{AreaID: areas[26].ID, PartyID: parties[0].ID, CandidateNo: 7, FullName: "นายนฤพล เลิศปัญญาโรจน์", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนายณัฐชา"},
		{AreaID: areas[27].ID, PartyID: parties[0].ID, CandidateNo: 3, FullName: "น.ส.ชลณัฏฐ์ โกยกุล", Biography: "ผู้สมัครหน้าใหม่ ลงแทนนางสาวรักชนก (ไอซ์)"},
		{AreaID: areas[28].ID, PartyID: parties[0].ID, CandidateNo: 3, FullName: "น.ส.ทิสรัตน์ เลาหพล", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[29].ID, PartyID: parties[0].ID, CandidateNo: 15, FullName: "นายธัญธร ธนินวัฒนาธร", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[30].ID, PartyID: parties[0].ID, CandidateNo: 12, FullName: "นายอนุสรณ์ ธรรมใจ", Biography: "ผู้สมัครหน้าใหม่ นักวิชาการเศรษฐศาสตร์"},
		{AreaID: areas[31].ID, PartyID: parties[0].ID, CandidateNo: 15, FullName: "น.ส.ปวิตรา จิตตกิจ", Biography: "อดีต สส. ป้องกันแชมป์"},
		{AreaID: areas[32].ID, PartyID: parties[0].ID, CandidateNo: 11, FullName: "นายเท่าพิภพ ลิ้มจิตรกร", Biography: "ขยับจากเขต 24 มาแก้ปัญหาวิกฤตฉุกเฉิน"},

		// ==========================================
		// พรรคเพื่อไทย (PartyID: parties[1].ID) - เขตยุทธศาสตร์
		// ==========================================
		{AreaID: areas[0].ID, PartyID: parties[1].ID, CandidateNo: 8, FullName: "นายญาณกิตติ์ ห่วงทรัพย์", Biography: "การพยายามปักธงในพื้นที่อนุรักษ์นิยม"},
		{AreaID: areas[1].ID, PartyID: parties[1].ID, CandidateNo: 5, FullName: "ดร.เดวิด มกรพงศ์", Biography: "เจาะกลุ่มผู้มีรายได้สูงย่านสาทร"},
		{AreaID: areas[2].ID, PartyID: parties[1].ID, CandidateNo: 2, FullName: "นางสาวเพ็ญพิสุทธิ์ จินตโสภณ", Biography: "การสืบทอดฐานเสียงทางการเมืองของตระกูลจินตโสภณในยานนาวา"},
		{AreaID: areas[3].ID, PartyID: parties[1].ID, CandidateNo: 4, FullName: "นางสาวบุณยกร ดำรงรัตน์", Biography: "ตัวแทนคนรุ่นใหม่ที่ถูกส่งมาเจาะฐานพื้นที่ชั้นใน"},
		{AreaID: areas[4].ID, PartyID: parties[1].ID, CandidateNo: 13, FullName: "นายขจรศักดิ์ ประดิษฐาน", Biography: "การรักษาฐานที่มั่นในเขตห้วยขวาง-วังทองหลาง"},
		{AreaID: areas[5].ID, PartyID: parties[1].ID, CandidateNo: 4, FullName: "นายสหัสวรรษ วีระมงคลกุล", Biography: "การเจาะกลุ่มวัยรุ่นและวัยทำงานตอนต้นในย่านพญาไท"},
		{AreaID: areas[6].ID, PartyID: parties[1].ID, CandidateNo: 3, FullName: "นายพุฒิพงศ์ อินทรสุวรรณ", Biography: "แข่งขันในเขตหน่วยงานทหารและข้าราชการ"},
		{AreaID: areas[7].ID, PartyID: parties[1].ID, CandidateNo: 5, FullName: "นายสุรชาติ เทียนทอง", Biography: "ใช้ความแข็งแกร่งของตระกูลเทียนทองเพื่อรักษาพื้นที่หลักสี่"},
		{AreaID: areas[8].ID, PartyID: parties[1].ID, CandidateNo: 8, FullName: "นายสายันต์ จันทร์เหมือนเผือก", Biography: "ระดมสรรพกำลังระดับรากหญ้าในพื้นที่บางเขน-จตุจักร"},
		{AreaID: areas[10].ID, PartyID: parties[1].ID, CandidateNo: 7, FullName: "นางสาวรัตติกาล แก้วเกิดมี", Biography: "ผู้ท้าชิงในเขตสายไหมที่เผชิญหน้ากับทั้งบ้านใหญ่และภาคประชาสังคม"},
		{AreaID: areas[32].ID, PartyID: parties[1].ID, CandidateNo: 1, FullName: "นายสุไพรพล เพ็ญแข", Biography: "โอกาสทองจากการสะดุดล้มของพรรคคู่แข่งในเขตบางพลัด"},

		// ==========================================
		// พรรคภูมิใจไทย (PartyID: parties[2].ID)
		// ==========================================
		{AreaID: areas[0].ID, PartyID: parties[2].ID, CandidateNo: 1, FullName: "น.ส.ลลิดา เพริศวิวัฒนา", Biography: ""},
		{AreaID: areas[1].ID, PartyID: parties[2].ID, CandidateNo: 4, FullName: "น.ส.พัชรินทร์ ซำศิริพงษ์", Biography: ""},
		{AreaID: areas[2].ID, PartyID: parties[2].ID, CandidateNo: 12, FullName: "นายสาโรช ต่อเทียนชัย", Biography: ""},
		{AreaID: areas[3].ID, PartyID: parties[2].ID, CandidateNo: 6, FullName: "นายเขตรัฐ เหล่าธรรมทัศน์", Biography: "อดีตนักการเมืองสายอนุรักษ์นิยม"},
		{AreaID: areas[4].ID, PartyID: parties[2].ID, CandidateNo: 4, FullName: "นายประเดิมชัย บุญช่วยเหลือ", Biography: ""},
		{AreaID: areas[5].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "นายนรเสฏฐ์ เธียรประสิทธิ์", Biography: ""},
		{AreaID: areas[6].ID, PartyID: parties[2].ID, CandidateNo: 6, FullName: "นางลลิตา ฤกษ์สำราญ", Biography: ""},
		{AreaID: areas[7].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "นายฤกษ์อารี นานา", Biography: ""},
		{AreaID: areas[8].ID, PartyID: parties[2].ID, CandidateNo: 6, FullName: "น.ส.ณัฐวรินธร บวรภัควุฒิสิริ", Biography: ""},
		{AreaID: areas[9].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "นายรณกร เชียรวิชัย", Biography: ""},
		{AreaID: areas[10].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "นายเอกภพ เหลืองประเสริฐ", Biography: "แกนนำภาคประชาสังคม สายไหมต้องรอด"},
		{AreaID: areas[11].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "นางศลิษา สิงหเสนี", Biography: ""},
		{AreaID: areas[12].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายศุกล กุลสิงห์", Biography: ""},
		{AreaID: areas[13].ID, PartyID: parties[2].ID, CandidateNo: 13, FullName: "น.ส.ฐิติภัสร์ โชติเดชาชัยนันต์", Biography: ""},
		{AreaID: areas[14].ID, PartyID: parties[2].ID, CandidateNo: 5, FullName: "นายถนอม อ่อนเกตุพล", Biography: ""},
		{AreaID: areas[15].ID, PartyID: parties[2].ID, CandidateNo: 2, FullName: "นายณัฐนันท์ กัลยาศิริ", Biography: ""},
		{AreaID: areas[16].ID, PartyID: parties[2].ID, CandidateNo: 4, FullName: "นายสุขสันต์ แสงศรี", Biography: ""},
		{AreaID: areas[17].ID, PartyID: parties[2].ID, CandidateNo: 5, FullName: "นายรณชัย สังฆมิตรกุล", Biography: ""},
		{AreaID: areas[18].ID, PartyID: parties[2].ID, CandidateNo: 9, FullName: "น.ส.กาญจนา ภวัครานนท์", Biography: ""},
		{AreaID: areas[19].ID, PartyID: parties[2].ID, CandidateNo: 2, FullName: "นายธนสิทธิ์ เมธพันธุ์เมือง", Biography: ""},
		{AreaID: areas[20].ID, PartyID: parties[2].ID, CandidateNo: 4, FullName: "นายทวนชัย นิยมชาติ", Biography: ""},
		{AreaID: areas[21].ID, PartyID: parties[2].ID, CandidateNo: 8, FullName: "นายพงศ์พล ยอดเมืองเจริญ", Biography: ""},
		{AreaID: areas[22].ID, PartyID: parties[2].ID, CandidateNo: 1, FullName: "ดร.สกุลรัตน์ ทิพย์วรรณงาม", Biography: ""},
		{AreaID: areas[23].ID, PartyID: parties[2].ID, CandidateNo: 5, FullName: "น.ส.เจณิสตา เตซะโสภณมณี", Biography: ""},
		{AreaID: areas[24].ID, PartyID: parties[2].ID, CandidateNo: 3, FullName: "นายเจริญศักดิ์ มณีรัตนสุบรรณ", Biography: ""},
		{AreaID: areas[25].ID, PartyID: parties[2].ID, CandidateNo: 10, FullName: "นายโชติพิพัฒน์ เตชะโสภณมณี", Biography: ""},
		{AreaID: areas[26].ID, PartyID: parties[2].ID, CandidateNo: 5, FullName: "นายศิลปชัย บุญราย", Biography: ""},
		{AreaID: areas[27].ID, PartyID: parties[2].ID, CandidateNo: 1, FullName: "น.ส.มัญดา อัฐจินดา", Biography: ""},
		{AreaID: areas[28].ID, PartyID: parties[2].ID, CandidateNo: 5, FullName: "น.ส.ธัณยาการย์ เตชะพัฒน์สิริ", Biography: ""},
		{AreaID: areas[29].ID, PartyID: parties[2].ID, CandidateNo: 2, FullName: "ดร.อำพล ขำวิลัย", Biography: ""},
		{AreaID: areas[30].ID, PartyID: parties[2].ID, CandidateNo: 2, FullName: "นายธนพล ชื่นพาณิชยกุล", Biography: ""},
		{AreaID: areas[31].ID, PartyID: parties[2].ID, CandidateNo: 1, FullName: "นายหมวดตรี ศุภิกา พัฒน์ธนันภู", Biography: ""},
		{AreaID: areas[32].ID, PartyID: parties[2].ID, CandidateNo: 10, FullName: "นายอรรทิตย์ฌาณ คูหาเรืองรอง", Biography: ""},

		// ==========================================
		// พรรคกล้าธรรม (PartyID: parties[3].ID) - เขตยุทธศาสตร์
		// ==========================================
		{AreaID: areas[1].ID, PartyID: parties[3].ID, CandidateNo: 3, FullName: "นางสาววิลาสินี แป๊ะสมัน", Biography: "การใช้สายสัมพันธ์ทางครอบครัวเพื่อสร้างจุดยึดโยงในเขตเศรษฐกิจหลัก"},
		{AreaID: areas[2].ID, PartyID: parties[3].ID, CandidateNo: 5, FullName: "นายนรุตม์ชัย บุนนาค", Biography: "การส่งตัวแทนที่มีภาพลักษณ์นักวิชาการ/นักกฎหมายในเขตยานนาวา"},
		{AreaID: areas[3].ID, PartyID: parties[3].ID, CandidateNo: 9, FullName: "ว่าที่ร้อยตรีหญิง ปภัสราวรรณ ม่วงไหม", Biography: "นำเสนอผู้หญิงเก่งเพื่อดึงดูดชนชั้นกลางในเขตคลองเตย"},
		{AreaID: areas[14].ID, PartyID: parties[3].ID, CandidateNo: 7, FullName: "น.ส.ปวริศา คุณาวรนนท์", Biography: "การเจาะกลุ่มฐานรากและชุมชนหมู่บ้านจัดสรรในเขตคันนายาว"},
		{AreaID: areas[21].ID, PartyID: parties[3].ID, CandidateNo: 10, FullName: "นายชณทัต รินน์นพคุณ", Biography: "แทรกซึมในพื้นที่สวนหลวง-ประเวศด้วยภาพลักษณ์นักบริหาร"},
		{AreaID: areas[22].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นายณัฐธนินทร์ เลิศเตชะสกุล", Biography: "การแข่งขันกับผู้สมัครหน้าใหม่ในเขตบางนา"},
		{AreaID: areas[25].ID, PartyID: parties[3].ID, CandidateNo: 9, FullName: "นายเสฐียรพงษ์ สำแดงสุข", Biography: "มุ่งเป้าไปที่พื้นที่ชานเมืองฝั่งธนบุรี (บางขุนเทียน-จอมทอง)"},
		{AreaID: areas[26].ID, PartyID: parties[3].ID, CandidateNo: 4, FullName: "นายมณูเทพย์ ทองปน", Biography: "การตัดคะแนนความนิยมจากฐานเสียงอนุรักษ์นิยมเดิมในพื้นที่บางบอน"},
		{AreaID: areas[29].ID, PartyID: parties[3].ID, CandidateNo: 1, FullName: "นายเอกรัฐ อิทธิไกวัล", Biography: "ผู้สมัครที่มีความเชี่ยวชาญเฉพาะเพื่อแข่งขันในเขตภาษีเจริญ"},
		{AreaID: areas[30].ID, PartyID: parties[3].ID, CandidateNo: 10, FullName: "น.ส.รักตาภา วงษ์ยอด", Biography: "การใช้ผู้สมัครสตรีเพื่อสร้างความแตกต่างในการหาเสียงย่านตลิ่งชัน"},
		{AreaID: areas[32].ID, PartyID: parties[3].ID, CandidateNo: 9, FullName: "นายกฤษณ์ สุริยผล", Biography: "การฉวยจังหวะความวุ่นวายของคู่แข่งเพื่อสร้างคะแนนเสียงในพื้นที่บางพลัด"},
	}

	if err := db.Create(&candidates).Error; err != nil {
		log.Fatalf("❌ Failed to seed candidates: %v", err)
	}

	log.Printf("✅ Seeded %d candidates successfully", len(candidates))
	return candidates
}
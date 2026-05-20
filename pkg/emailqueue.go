package pkg

import (
	"log"
)

type emailJob struct {
	to      string
	otp     string
	refCode string
}

var emailQueue chan emailJob

// StartEmailWorker เริ่ม worker pool สำหรับส่ง email แบบ async
// เรียกครั้งเดียวตอน server start
func StartEmailWorker(workers int) {
	emailQueue = make(chan emailJob, 100)
	for i := 0; i < workers; i++ {
		go func() {
			for job := range emailQueue {
				if err := SendOTPEmail(job.to, job.otp, job.refCode); err != nil {
					log.Printf("async email failed to=%s ref=%s: %v", job.to, job.refCode, err)
				}
			}
		}()
	}
	log.Printf("Email worker pool started (%d workers)", workers)
}

// EnqueueOTPEmail ส่ง email แบบ async — non-blocking
// ถ้า queue เต็ม (>100 jobs) จะ fallback เป็น sync
func EnqueueOTPEmail(to, otp, refCode string) error {
	if emailQueue == nil {
		return SendOTPEmail(to, otp, refCode)
	}
	select {
	case emailQueue <- emailJob{to: to, otp: otp, refCode: refCode}:
		return nil
	default:
		// queue เต็ม → sync fallback
		log.Println("email queue full, sending synchronously")
		return SendOTPEmail(to, otp, refCode)
	}
}

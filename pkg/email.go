package pkg

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

var SMTPSendMail func(addr string, a smtp.Auth, from string, to []string, msg []byte) error = smtp.SendMail

// SendOTPEmail ส่ง OTP ไปยัง email ของผู้ใช้ผ่าน SMTP
// ถ้าไม่ได้ตั้ง SMTP_HOST จะ skip (dev mode)
func SendOTPEmail(toEmail, otpCode, refCode string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if host == "" {
		log.Println("SMTP_HOST not set — skipping email send (dev mode)")
		return nil
	}
	if port == "" {
		port = "587"
	}
	if from == "" {
		from = user
	}
	if user == "" || password == "" {
		return fmt.Errorf("SMTP_USER และ SMTP_PASSWORD ต้องตั้งค่าก่อนใช้งาน")
	}

	auth := smtp.PlainAuth("", user, password, host)
	htmlBody := buildOTPEmailHTML(otpCode, refCode)

	msg := fmt.Sprintf(
		"From: VoteSpher <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: รหัส OTP สำหรับการยืนยันตัวตน — VoteSpher\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		from, toEmail, htmlBody,
	)

	addr := fmt.Sprintf("%s:%s", host, port)
	if err := SMTPSendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("ส่ง email ไม่สำเร็จ: %w", err)
	}

	log.Printf("OTP email sent to %s (ref: %s)", toEmail, refCode)
	return nil
}

func buildOTPEmailHTML(otpCode, refCode string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="th">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#eef2f7;font-family:'Segoe UI',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#eef2f7;padding:48px 16px;">
    <tr>
      <td align="center">
        <table width="520" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">

          <!-- Header -->
          <tr>
            <td style="background:linear-gradient(135deg,#1e3a8a 0%%,#1d4ed8 100%%);padding:36px 48px;text-align:center;">
              <div style="display:inline-block;background:rgba(255,255,255,0.15);border-radius:50%%;padding:14px 18px;margin-bottom:12px;">
                <span style="font-size:28px;">&#128499;</span>
              </div>
              <div style="color:#ffffff;font-size:22px;font-weight:700;letter-spacing:1px;">VoteSpher</div>
              <div style="color:#93c5fd;font-size:13px;margin-top:4px;">ระบบการเลือกตั้งออนไลน์</div>
            </td>
          </tr>

          <!-- Body -->
          <tr>
            <td style="padding:40px 48px 32px;">

              <p style="color:#1e293b;font-size:16px;font-weight:600;margin:0 0 8px;">สวัสดีครับ/ค่ะ</p>
              <p style="color:#64748b;font-size:14px;line-height:1.7;margin:0 0 32px;">
                คุณได้ร้องขอรหัส OTP เพื่อยืนยันตัวตนในระบบ VoteSpher
                กรุณานำรหัสด้านล่างไปกรอกในแอปพลิเคชันให้ครบถ้วน
              </p>

              <!-- OTP Box -->
              <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:20px;">
                <tr>
                  <td align="center" style="background:#eff6ff;border:2px dashed #3b82f6;border-radius:12px;padding:28px 16px;">
                    <div style="color:#64748b;font-size:11px;letter-spacing:2px;text-transform:uppercase;margin-bottom:10px;font-weight:600;">รหัส OTP ของคุณ</div>
                    <div style="color:#1e3a8a;font-size:44px;font-weight:800;letter-spacing:16px;font-family:'Courier New',monospace;">%s</div>
                  </td>
                </tr>
              </table>

              <!-- Ref Code -->
              <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
                <tr>
                  <td align="center" style="background:#f8fafc;border-radius:8px;padding:12px;">
                    <span style="color:#94a3b8;font-size:13px;">Ref Code:&nbsp;</span>
                    <span style="color:#334155;font-size:13px;font-weight:700;font-family:'Courier New',monospace;letter-spacing:1px;">%s</span>
                  </td>
                </tr>
              </table>

              <!-- Warning -->
              <table width="100%%" cellpadding="0" cellspacing="0" style="margin-bottom:28px;">
                <tr>
                  <td style="background:#fefce8;border-left:4px solid #eab308;border-radius:0 8px 8px 0;padding:14px 16px;">
                    <p style="color:#854d0e;font-size:13px;margin:0;line-height:1.6;">
                      <strong>&#9888;&#65039; โปรดทราบ:</strong><br>
                      รหัสนี้มีอายุ <strong>5 นาที</strong> และใช้ได้เพียงครั้งเดียวเท่านั้น<br>
                      ทีมงาน VoteSpher จะ<strong>ไม่ขอรหัสนี้</strong>จากคุณทางช่องทางใดทั้งสิ้น
                    </p>
                  </td>
                </tr>
              </table>

              <p style="color:#94a3b8;font-size:13px;margin:0;line-height:1.6;">
                หากคุณไม่ได้เป็นผู้ร้องขอรหัสนี้ กรุณาเพิกเฉยต่ออีเมลฉบับนี้ได้เลย
              </p>
            </td>
          </tr>

          <!-- Divider -->
          <tr>
            <td style="padding:0 48px;"><hr style="border:none;border-top:1px solid #e2e8f0;margin:0;"></td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding:20px 48px;text-align:center;">
              <p style="color:#cbd5e1;font-size:12px;margin:0;">
                &copy; 2026 VoteSpher &mdash; ส่งโดยอัตโนมัติ กรุณาอย่าตอบกลับอีเมลนี้
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, otpCode, refCode)
}

// Package mail يرسل إيميلات النظام (رموز التحقق لتطبيق الزبائن) عن طريق SMTP.
//
// قرار (ع) 10-14: Gmail بكلمة سر تطبيق (App Password). بلا إعدادات SMTP
// (التطوير المحلي) الرسالة تنكتب باللوك بدل ما تنرسل.
package mail

import (
	"crypto/tls"
	"fmt"
	"html"
	"log"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Sender struct {
	host, port, user, pass, from string
}

func FromEnv() *Sender {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("MAIL_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USER")
	}
	return &Sender{host: os.Getenv("SMTP_HOST"), port: port, user: os.Getenv("SMTP_USER"), pass: os.Getenv("SMTP_PASS"), from: from}
}

func (s *Sender) Configured() bool { return s.host != "" && s.user != "" && s.pass != "" }

// Send يرسل HTML. بلا إعدادات يطبع الرسالة باللوك (للتطوير).
func (s *Sender) Send(to, subject, body string) error {
	if !s.Configured() {
		log.Printf("[mail:dev] to=%s subject=%s\n%s", to, subject, body)
		return nil
	}
	msg := strings.Join([]string{
		"From: " + mime.QEncoding.Encode("utf-8", "شركة الأماني") + " <" + s.from + ">",
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Date: " + time.Now().Format(time.RFC1123Z),
		"", body,
	}, "\r\n")
	addr := net.JoinHostPort(s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	if s.port == "465" {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
		if err != nil {
			return err
		}
		c, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return err
		}
		defer c.Close()
		if err := c.Auth(auth); err != nil {
			return err
		}
		if err := c.Mail(s.from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		wc, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := wc.Write([]byte(msg)); err != nil {
			return err
		}
		return wc.Close()
	}
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// OtpHTML قالب رسالة الرمز بالعربي.
func OtpHTML(name, code, purpose string) string {
	what := "لتأكيد حسابك بتطبيق الأماني"
	if purpose == "reset" {
		what = "لتغيير كلمة المرور بتطبيق الأماني"
	}
	return fmt.Sprintf(`<div dir="rtl" style="font-family:Tahoma,Arial;max-width:480px;margin:auto;padding:24px;border:1px solid #e2e8f0;border-radius:16px">
<h2 style="color:#1e3a8a;margin:0 0 8px">أهلاً %s 👋</h2>
<p style="color:#334155">هذا رمز التحقق %s:</p>
<p style="font-size:34px;letter-spacing:10px;font-weight:bold;color:#0f172a;text-align:center;background:#f1f5f9;border-radius:12px;padding:14px">%s</p>
<p style="color:#64748b;font-size:13px">الرمز صالح لمدة 10 دقائق. إذا ما طلبته أنت، تجاهل هذه الرسالة.</p>
<p style="color:#94a3b8;font-size:12px">شركة الأماني</p></div>`, html.EscapeString(name), what, code)
}

package notifier

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Notifier struct {
	smsLogin     string
	smsApiKey    string
	smsSender    string
	smsApiURL    string
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPass     string
	smtpFrom     string
	emailTargets []string
}

func NewNotifier() *Notifier {
	// По умолчанию используем 587 (STARTTLS)
	smtpPort := 587
	if val := os.Getenv("SMTP_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &smtpPort)
	}

	emailTargets := strings.Split(os.Getenv("EMAIL_OPERATORS"), ",")
	for i, v := range emailTargets {
		emailTargets[i] = strings.TrimSpace(v)
	}

	return &Notifier{
		smsLogin:     os.Getenv("SMS_LOGIN"),
		smsApiKey:    os.Getenv("SMS_API_KEY"),
		smsSender:    os.Getenv("SMS_SENDER_NAME"),
		smsApiURL:    os.Getenv("REDSMS_API_URL"),
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     smtpPort,
		smtpUser:     os.Getenv("SMTP_USER"),
		smtpPass:     os.Getenv("SMTP_PASS"),
		smtpFrom:     os.Getenv("SMTP_FROM"),
		emailTargets: emailTargets,
	}
}

func (n *Notifier) SendCodeBySms(phone string, code string) error {
	text := fmt.Sprintf("Код подтверждения: %s. Emelia Invest", code)
	return n.sendSms(phone, text)
}

func (n *Notifier) SendLoginAndPasswordBySms(phone string, login string, password string) error {
	text := fmt.Sprintf("Логин: %s\nПароль: %s\nEmelia Invest", login, password)
	return n.sendSms(phone, text)
}

func (n *Notifier) sendSms(phone string, text string) error {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	hash := md5.Sum([]byte(ts + n.smsApiKey))
	secret := hex.EncodeToString(hash[:])

	payload := map[string]interface{}{
		"route": "sms",
		"from":  n.smsSender,
		"to":    phone,
		"text":  text,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", n.smsApiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("login", n.smsLogin)
	req.Header.Set("ts", ts)
	req.Header.Set("secret", secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка ответа RedSMS: %s", resp.Status)
	}

	log.Println("[NOTIFIER] SMS отправлено:", phone)
	return nil
}

func (n *Notifier) SendEmailToOperator(subject, body string) error {
	if n.smtpHost == "" || n.smtpUser == "" || n.smtpPass == "" || n.smtpPort == 0 {
		return fmt.Errorf("[NOTIFIER] SMTP env не заданы (SMTP_HOST/PORT/USER/PASS)")
	}

	log.Println("[NOTIFIER] Отправка email через SMTP (STARTTLS)")

	tryFrom := "no-reply@emelia-invest.com"
	if err := n.sendEmailsWithFrom(tryFrom, subject, body); err != nil {
		log.Printf("[NOTIFIER] Не удалось отправить от имени %s: %v. Пробуем из ENV...", tryFrom, err)
		if n.smtpFrom == "" {
			return fmt.Errorf("[NOTIFIER] SMTP_FROM не задан для fallback")
		}
		return n.sendEmailsWithFrom(n.smtpFrom, subject, body)
	}
	return nil
}

// sendEmailsWithFrom — выделенная функция, чтобы не дублировать код отправки
func (n *Notifier) sendEmailsWithFrom(from string, subject, body string) error {
	baseHeaders := map[string]string{
		"From":         from,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	addr := fmt.Sprintf("%s:%d", n.smtpHost, n.smtpPort)
	tlsCfg := &tls.Config{ServerName: n.smtpHost}

	for _, to := range n.emailTargets {
		to = strings.TrimSpace(to)
		if to == "" {
			continue
		}

		var sb strings.Builder
		for k, v := range baseHeaders {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
		sb.WriteString("\r\n")
		sb.WriteString(body)
		msg := []byte(sb.String())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("[NOTIFIER] SMTP dial: %w", err)
		}

		c, err := smtp.NewClient(conn, n.smtpHost)
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("[NOTIFIER] SMTP new client: %w", err)
		}

		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsCfg); err != nil {
				_ = c.Close()
				return fmt.Errorf("[NOTIFIER] STARTTLS: %w", err)
			}
		}

		auth := smtp.PlainAuth("", n.smtpUser, n.smtpPass, n.smtpHost)
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				_ = c.Close()
				return fmt.Errorf("[NOTIFIER] SMTP auth: %w", err)
			}
		}

		// envelope-from можно оставить как smtpUser — так безопаснее для SPF
		if err := c.Mail(n.smtpUser); err != nil {
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] MAIL FROM: %w", err)
		}
		if err := c.Rcpt(to); err != nil {
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] RCPT TO (%s): %w", to, err)
		}

		w, err := c.Data()
		if err != nil {
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] DATA open: %w", err)
		}
		if _, err := w.Write(msg); err != nil {
			_ = w.Close()
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] DATA write: %w", err)
		}
		if err := w.Close(); err != nil {
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] DATA close: %w", err)
		}

		if err := c.Quit(); err != nil {
			_ = c.Close()
			return fmt.Errorf("[NOTIFIER] SMTP quit: %w", err)
		}

		log.Printf("[NOTIFIER] Email отправлен от %s: %s\n", from, to)
	}
	return nil
}

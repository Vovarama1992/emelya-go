package notifier

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/gomail.v2"
)

type Notifier struct {
	smsLogin  string
	smsApiKey string
	smsSender string
	smtpHost  string
	smtpPort  int
	smtpUser  string
	smtpPass  string
	smtpFrom  string
}

func NewNotifier() *Notifier {
	return &Notifier{
		smsLogin:  os.Getenv("SMS_LOGIN"),
		smsApiKey: os.Getenv("SMS_API_KEY"),
		smsSender: os.Getenv("SMS_SENDER_NAME"),
		smtpHost:  os.Getenv("SMTP_HOST"),
		smtpPort:  465,
		smtpUser:  os.Getenv("SMTP_USER"),
		smtpPass:  os.Getenv("SMTP_PASS"),
		smtpFrom:  os.Getenv("SMTP_FROM"),
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

	req, err := http.NewRequest("POST", "https://cp.redsms.ru/api/message", bytes.NewBuffer(body))
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
	log.Println("[NOTIFIER] Начало отправки email оператору")

	m := gomail.NewMessage()
	from := n.smtpFrom
	to := []string{"vital80@inbox.ru", "vovayhh9988@gmail.com"}

	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	log.Printf("[NOTIFIER] Email параметры:\n  From: %s\n  To: %v\n  Subject: %s\n  Body: %s",
		from, to, subject, body)

	d := gomail.NewDialer(n.smtpHost, n.smtpPort, n.smtpUser, n.smtpPass)
	d.SSL = true

	log.Printf("[NOTIFIER] SMTP настройки:\n  Host: %s\n  Port: %d\n  User: %s\n  SSL: %v",
		n.smtpHost, n.smtpPort, n.smtpUser, d.SSL)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("[NOTIFIER: EMAIL] Ошибка отправки: %v", err)
		return err
	}

	log.Println("[NOTIFIER] Email успешно отправлен операторам.")
	return nil
}

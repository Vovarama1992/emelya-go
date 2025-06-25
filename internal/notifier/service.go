package notifier

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Notifier struct {
	smsLogin    string
	smsApiKey   string
	smsSender   string
	emailApiKey string
	smtpHost    string
	smtpPort    int
	smtpUser    string
	smtpPass    string
	smtpFrom    string
}

func NewNotifier() *Notifier {
	return &Notifier{
		smsLogin:    os.Getenv("SMS_LOGIN"),
		smsApiKey:   os.Getenv("SMS_API_KEY"),
		smsSender:   os.Getenv("SMS_SENDER_NAME"),
		smtpHost:    os.Getenv("SMTP_HOST"),
		smtpPort:    465,
		emailApiKey: os.Getenv("EMAIL_API_KEY"),
		smtpUser:    os.Getenv("SMTP_USER"),
		smtpPass:    os.Getenv("SMTP_PASS"),
		smtpFrom:    os.Getenv("SMTP_FROM"),
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

	apiKey := n.emailApiKey
	from := n.smtpFrom
	toList := []string{"vital80@inbox.ru", "emelyainvest@gmail.com", "emelya.tech@mail.ru"}

	type EmailRequest struct {
		FromEmail string `json:"from_email"`
		Subject   string `json:"subject"`
		Text      string `json:"text"`
		To        string `json:"to"`
	}

	client := &http.Client{}
	for _, to := range toList {
		reqBody := EmailRequest{
			FromEmail: from,
			Subject:   subject,
			Text:      body,
			To:        to,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("[NOTIFIER: EMAIL] Ошибка маршалинга JSON: %v", err)
			return err
		}

		req, err := http.NewRequest("POST", "https://api.msndr.net/v1/email/messages", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[NOTIFIER: EMAIL] Ошибка создания запроса: %v", err)
			return err
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[NOTIFIER: EMAIL] Ошибка отправки запроса: %v", err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("[NOTIFIER: EMAIL] Ошибка ответа API: %s", string(bodyBytes))
			return fmt.Errorf("не удалось отправить email, статус: %s", resp.Status)
		}
	}

	log.Println("[NOTIFIER] Email успешно отправлен операторам через RedSMS API.")
	return nil
}

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
	"strings"
	"time"
)

type Notifier struct {
	smsLogin     string
	smsApiKey    string
	smsSender    string
	smsApiURL    string
	emailApiKey  string
	emailApiURL  string
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPass     string
	smtpFrom     string
	emailTargets []string
}

func NewNotifier() *Notifier {
	smtpPort := 465
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
		emailApiKey:  os.Getenv("EMAIL_API_KEY"),
		emailApiURL:  os.Getenv("EMAIL_API_URL"),
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
	log.Println("[NOTIFIER] Начало отправки email оператору")

	type EmailRequest struct {
		FromEmail string `json:"from_email"`
		Subject   string `json:"subject"`
		Text      string `json:"text"`
		To        string `json:"to"`
	}

	client := &http.Client{}
	for _, to := range n.emailTargets {
		reqBody := EmailRequest{
			FromEmail: n.smtpFrom,
			Subject:   subject,
			Text:      body,
			To:        strings.TrimSpace(to),
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("[NOTIFIER: EMAIL] Ошибка маршалинга JSON: %v", err)
			return err
		}

		req, err := http.NewRequest("POST", n.emailApiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[NOTIFIER: EMAIL] Ошибка создания запроса: %v", err)
			return err
		}
		req.Header.Set("Authorization", "Bearer "+n.emailApiKey)
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

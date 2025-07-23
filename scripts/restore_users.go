package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DumpUser struct {
	ID              int64    `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Patronymic      *string  `json:"patronymic"`
	Email           string   `json:"email"`
	Phone           string   `json:"phone"`
	Login           string   `json:"login"`
	PasswordHash    string   `json:"password_hash"`
	ReferrerID      *float64 `json:"referrer_id"`
	CardNumber      *string  `json:"card_number"`
	Balance         string   `json:"balance"`
	IsEmailVerified bool     `json:"is_email_verified"`
	IsPhoneVerified bool     `json:"is_phone_verified"`
}

func getContainerIP(serviceName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", serviceName)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	ip, err := getContainerIP("db")
	if err != nil {
		log.Fatalf("Ошибка получения IP контейнера: %v", err)
	}

	dsn := fmt.Sprintf("postgres://emelya:secret@%s:5432/emelya_db?sslmode=disable", ip)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	file, err := os.Open("users_from_dump.json")
	if err != nil {
		log.Fatal("Не удалось открыть users_from_dump.json:", err)
	}
	defer file.Close()

	var users []DumpUser
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Fatal("Ошибка парсинга JSON:", err)
	}

	ctx := context.Background()
	for _, u := range users {
		var refID *int64
		if u.ReferrerID != nil {
			r := int64(*u.ReferrerID)
			refID = &r
		}

		var insertedID int64
		err := db.GetContext(ctx, &insertedID, `
			INSERT INTO users (
				id, first_name, last_name, patronymic, email, phone,
				is_email_verified, is_phone_verified,
				login, password_hash, referrer_id, card_number
			) VALUES (
				:id, :first_name, :last_name, :patronymic, :email, :phone,
				:is_email_verified, :is_phone_verified,
				:login, :password_hash, :referrer_id, :card_number
			)
			RETURNING id
		`, map[string]interface{}{
			"id":                u.ID,
			"first_name":        u.FirstName,
			"last_name":         u.LastName,
			"patronymic":        u.Patronymic,
			"email":             u.Email,
			"phone":             u.Phone,
			"is_email_verified": u.IsEmailVerified,
			"is_phone_verified": u.IsPhoneVerified,
			"login":             u.Login,
			"password_hash":     u.PasswordHash,
			"referrer_id":       refID,
			"card_number":       u.CardNumber,
		})
		if err != nil {
			log.Fatalf("Ошибка вставки пользователя %s: %v", u.Email, err)
		}

		if strings.TrimSpace(u.Balance) != "0.00" && u.Balance != "" {
			_, err := db.ExecContext(ctx, `
				INSERT INTO deposits (
					user_id, amount, created_at, approved_at,
					block_until, daily_reward, status
				) VALUES ($1, $2, now(), now(), NULL, NULL, 'approved')
			`, u.ID, u.Balance)
			if err != nil {
				log.Fatalf("Ошибка вставки депозита для %s: %v", u.Email, err)
			}
		}

		fmt.Printf("✅ %s inserted (id=%d)\n", u.Email, u.ID)
	}
}

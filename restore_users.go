package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RawUser struct {
	ID              int64           `json:"id"`
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	Patronymic      *string         `json:"patronymic"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	IsEmailVerified bool            `json:"is_email_verified"`
	IsPhoneVerified bool            `json:"is_phone_verified"`
	Login           string          `json:"login"`
	PasswordHash    string          `json:"password_hash"`
	ReferrerID      *int64          `json:"referrer_id"`
	CardNumber      *string         `json:"card_number"`
	BalanceRaw      json.RawMessage `json:"balance"`
}

func parseBalance(b json.RawMessage) float64 {
	if len(b) == 0 || string(b) == "null" {
		return 0
	}
	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		return f
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f, _ = strconv.ParseFloat(s, 64)
		return f
	}
	return 0
}

func getContainerIP(containerName string) (string, error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w", err)
	}
	ip := strings.TrimSpace(out.String())
	if ip == "" {
		return "", fmt.Errorf("empty IP from inspect")
	}
	return ip, nil
}

func main() {
	ip, err := getContainerIP("back-db-1")
	if err != nil {
		log.Fatalf("Ошибка получения IP контейнера: %v", err)
	}

	dsn := fmt.Sprintf("postgres://emelya:secret@%s:5432/emelya_db?sslmode=disable", ip)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	data, err := os.ReadFile("scripts/users_from_dump.json")
	if err != nil {
		log.Fatalf("Не удалось прочитать файл: %v", err)
	}

	var users []RawUser
	if err := json.Unmarshal(data, &users); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	ctx := context.Background()

	for _, u := range users {
		balance := parseBalance(u.BalanceRaw)

		_, err := db.NamedExecContext(ctx, `
			INSERT INTO users (
				id, first_name, last_name, patronymic, email, phone,
				is_email_verified, is_phone_verified, login, password_hash,
				referrer_id, card_number
			) OVERRIDING SYSTEM VALUE VALUES (
				:id, :first_name, :last_name, :patronymic, :email, :phone,
				:is_email_verified, :is_phone_verified, :login, :password_hash,
				:referrer_id, :card_number
			)
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
			"referrer_id":       u.ReferrerID,
			"card_number":       u.CardNumber,
		})
		if err != nil {
			log.Fatalf("Ошибка вставки пользователя %s: %v", u.Email, err)
		}

		if balance > 0 {
			_, err = db.ExecContext(ctx, `
				INSERT INTO deposits (
					user_id, amount, created_at, approved_at,
					block_until, daily_reward, status
				) VALUES ($1, $2, now(), now(), NULL, NULL, 'approved')
			`, u.ID, balance)
			if err != nil {
				log.Fatalf("Ошибка создания депозита для user_id=%d: %v", u.ID, err)
			}
		}

		fmt.Printf("Пользователь %s вставлен (id=%d, баланс=%.2f)\n", u.Email, u.ID, balance)
	}
}

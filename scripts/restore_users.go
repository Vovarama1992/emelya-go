package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

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

func main() {
	// ВНИМАНИЕ: db -> localhost, если запускаешь не из докера
	const dsn = "postgres://emelya:secret@localhost:5432/emelya_db?sslmode=disable"

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	dumpPath := filepath.Join("scripts", "users_from_dump.json")
	file, err := os.Open(dumpPath)
	if err != nil {
		log.Fatalf("Не удалось открыть JSON: %v", err)
	}
	defer file.Close()

	var users []DumpUser
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Fatalf("Ошибка при чтении JSON: %v", err)
	}

	ctx := context.Background()

	for _, u := range users {
		// referrer
		var refID *int64
		if u.ReferrerID != nil {
			id := int64(*u.ReferrerID)
			refID = &id
		}

		// balance
		balance, err := strconv.ParseFloat(u.Balance, 64)
		if err != nil {
			log.Fatalf("Невалидный баланс у пользователя %v: %v", u.ID, err)
		}

		// вставка юзера
		_, err = db.NamedExecContext(ctx, `
			INSERT INTO users (
				id, first_name, last_name, patronymic, email, phone,
				login, password_hash, referrer_id, card_number,
				is_email_verified, is_phone_verified
			) VALUES (
				:id, :first_name, :last_name, :patronymic, :email, :phone,
				:login, :password_hash, :referrer_id, :card_number,
				:is_email_verified, :is_phone_verified
			)
		`, map[string]interface{}{
			"id":                u.ID,
			"first_name":        u.FirstName,
			"last_name":         u.LastName,
			"patronymic":        u.Patronymic,
			"email":             u.Email,
			"phone":             u.Phone,
			"login":             u.Login,
			"password_hash":     u.PasswordHash,
			"referrer_id":       refID,
			"card_number":       u.CardNumber,
			"is_email_verified": u.IsEmailVerified,
			"is_phone_verified": u.IsPhoneVerified,
		})
		if err != nil {
			log.Fatalf("Ошибка вставки пользователя %v: %v", u.Email, err)
		}

		if balance > 0 {
			_, err = db.ExecContext(ctx, `
				INSERT INTO deposits (
					user_id, amount, created_at, approved_at,
					block_until, daily_reward, status
				) VALUES ($1, $2, now(), now(), NULL, NULL, 'approved')
			`, u.ID, balance)
			if err != nil {
				log.Fatalf("Ошибка вставки депозита для %v: %v", u.Email, err)
			}
		}

		fmt.Printf("✅ Вставлен пользователь %s с id %d\n", u.Email, u.ID)
	}
}

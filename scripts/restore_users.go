package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	PasswordHash    *string  `json:"password_hash"`
	ReferrerID      *float64 `json:"referrer_id"`
	CardNumber      *string  `json:"card_number"`
	BalanceRaw      string   `json:"balance"`
	IsEmailVerified bool     `json:"is_email_verified"`
	IsPhoneVerified bool     `json:"is_phone_verified"`
}

func main() {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer db.Close()

	file, err := os.Open("scripts/users_from_dump.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var users []DumpUser
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	for _, u := range users {
		balance, err := strconv.ParseFloat(u.BalanceRaw, 64)
		if err != nil {
			log.Fatalf("invalid balance for user %d: %v", u.ID, err)
		}

		params := map[string]interface{}{
			"id":                u.ID,
			"first_name":        u.FirstName,
			"last_name":         u.LastName,
			"patronymic":        u.Patronymic,
			"email":             u.Email,
			"phone":             u.Phone,
			"login":             u.Login,
			"password_hash":     defaultOrValue(u.PasswordHash, "$2a$10$changemechangemechangemeu"), // безопасный заглушка-хеш
			"referrer_id":       refIDToInt64(u.ReferrerID),
			"card_number":       u.CardNumber,
			"is_email_verified": u.IsEmailVerified,
			"is_phone_verified": u.IsPhoneVerified,
		}

		_, err = db.NamedExecContext(ctx, `
			INSERT INTO users (
				id, first_name, last_name, patronymic,
				email, phone, login, password_hash,
				referrer_id, card_number,
				is_email_verified, is_phone_verified
			) VALUES (
				:id, :first_name, :last_name, :patronymic,
				:email, :phone, :login, :password_hash,
				:referrer_id, :card_number,
				:is_email_verified, :is_phone_verified
			)
		`, params)
		if err != nil {
			log.Fatalf("failed to insert user %s: %v", u.Email, err)
		}

		if balance > 0 {
			_, err = db.ExecContext(ctx, `
				INSERT INTO deposits (
					user_id, amount, created_at, approved_at,
					block_until, daily_reward, status
				) VALUES ($1, $2, now(), now(), NULL, NULL, 'approved')
			`, u.ID, balance)
			if err != nil {
				log.Fatalf("failed to create deposit for user %d: %v", u.ID, err)
			}
		}

		fmt.Printf("User %s inserted (id %d, balance %.2f)\n", u.Email, u.ID, balance)
	}
}

func defaultOrValue(s *string, def string) string {
	if s == nil || *s == "" {
		return def
	}
	return *s
}

func refIDToInt64(f *float64) *int64 {
	if f == nil {
		return nil
	}
	v := int64(*f)
	return &v
}

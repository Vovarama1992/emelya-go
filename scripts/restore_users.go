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
	ID              int64           `json:"id"`
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	Patronymic      *string         `json:"patronymic"`
	Email           string          `json:"email"`
	Phone           string          `json:"phone"`
	Login           string          `json:"login"`
	PasswordHash    string          `json:"password_hash"`
	ReferrerID      *float64        `json:"referrer_id"`
	CardNumber      *string         `json:"card_number"`
	IsEmailVerified bool            `json:"is_email_verified"`
	IsPhoneVerified bool            `json:"is_phone_verified"`
	BalanceRaw      json.RawMessage `json:"balance"`
	Balance         float64         `json:"-"`
}

func getPostgresIP(containerName string) (string, error) {
	out, err := exec.Command("docker", "inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName).Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container IP: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	containerName := "back-db-1"
	pgIP, err := getPostgresIP(containerName)
	if err != nil {
		log.Fatalf("Ошибка получения IP контейнера: %v", err)
	}

	dsn := fmt.Sprintf("postgres://emelya:secret@%s:5432/emelya_db?sslmode=disable", pgIP)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	// Очистка таблиц
	if _, err := db.Exec(`DELETE FROM deposits`); err != nil {
		log.Fatalf("Ошибка очистки deposits: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM users`); err != nil {
		log.Fatalf("Ошибка очистки users: %v", err)
	}

	// Чтение файла
	file, err := os.Open("scripts/users_from_dump.json")
	if err != nil {
		log.Fatalf("Не удалось открыть users_from_dump.json: %v", err)
	}
	defer file.Close()

	var users []DumpUser
	if err := json.NewDecoder(file).Decode(&users); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	// Парсинг баланса
	for i := range users {
		var f float64
		if err := json.Unmarshal(users[i].BalanceRaw, &f); err != nil {
			var s string
			if err := json.Unmarshal(users[i].BalanceRaw, &s); err == nil {
				fmt.Sscanf(s, "%f", &f)
			}
		}
		users[i].Balance = f
	}

	ctx := context.Background()

	insertUserStmt, err := db.PrepareNamed(`
		INSERT INTO users (
			id, first_name, last_name, patronymic, email, phone,
			login, password_hash, card_number,
			is_email_verified, is_phone_verified
		)
		VALUES (
			:id, :first_name, :last_name, :patronymic, :email, :phone,
			:login, :password_hash, :card_number,
			:is_email_verified, :is_phone_verified
		)
		RETURNING id
	`)
	if err != nil {
		log.Fatalf("Ошибка подготовки INSERT-запроса: %v", err)
	}
	defer insertUserStmt.Close()

	for _, u := range users {
		// если такой email уже есть — удалим старую запись
		var exists bool
		err := db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, u.Email)
		if err != nil {
			log.Fatalf("Ошибка при проверке email %s: %v", u.Email, err)
		}
		if exists {
			_, err := db.ExecContext(ctx, `DELETE FROM users WHERE email = $1`, u.Email)
			if err != nil {
				log.Fatalf("Не удалось удалить дубликат email %s: %v", u.Email, err)
			}
			fmt.Printf("Старый пользователь с email %s удалён\n", u.Email)
		}

		params := map[string]interface{}{
			"id":                u.ID,
			"first_name":        u.FirstName,
			"last_name":         u.LastName,
			"patronymic":        u.Patronymic,
			"email":             u.Email,
			"phone":             u.Phone,
			"login":             u.Login,
			"password_hash":     u.PasswordHash,
			"card_number":       u.CardNumber,
			"is_email_verified": u.IsEmailVerified,
			"is_phone_verified": u.IsPhoneVerified,
		}

		var insertedID int64
		if err := insertUserStmt.GetContext(ctx, &insertedID, params); err != nil {
			log.Fatalf("Ошибка вставки пользователя %s: %v", u.Email, err)
		}

		if u.Balance > 0 {
			_, err := db.ExecContext(ctx, `
				INSERT INTO deposits (
					user_id, amount, created_at, approved_at,
					block_until, daily_reward, status
				) VALUES ($1, $2, now(), now(), NULL, NULL, 'approved')
			`, insertedID, u.Balance)
			if err != nil {
				log.Fatalf("Ошибка создания депозита для пользователя %d: %v", insertedID, err)
			}
		}

		fmt.Printf("Пользователь %s вставлен с id %d\n", u.Email, insertedID)
	}

	// Обновление referrer_id
	for _, u := range users {
		if u.ReferrerID != nil {
			ref := int64(*u.ReferrerID)
			_, err := db.ExecContext(ctx, `
				UPDATE users SET referrer_id = $1 WHERE id = $2
			`, ref, u.ID)
			if err != nil {
				log.Fatalf("Ошибка обновления referrer_id для пользователя %d: %v", u.ID, err)
			}
		}
	}
}

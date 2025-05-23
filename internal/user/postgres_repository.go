package user

import (
	"context"

	"github.com/Vovarama1992/emelya-go/internal/db"
)

type PostgresRepository struct {
	DB *db.DB
}

func NewPostgresRepository(db *db.DB) *PostgresRepository {
	return &PostgresRepository{
		DB: db,
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (
			first_name, last_name, patronymic, email, phone,
			is_email_verified, is_phone_verified, login, password_hash, referrer_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := r.DB.Pool.QueryRow(ctx, query,
		user.FirstName,
		user.LastName,
		user.Patronymic,
		user.Email,
		user.Phone,
		user.IsEmailVerified,
		user.IsPhoneVerified,
		user.Login,
		user.PasswordHash,
		user.ReferrerID, // может быть nil
	).Scan(&user.ID)

	return err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, userID int) (*User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified, login, password_hash, referrer_id
		FROM users
		WHERE id = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, userID)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
		&user.Email,
		&user.Phone,
		&user.IsEmailVerified,
		&user.IsPhoneVerified,
		&user.Login,
		&user.PasswordHash,
		&user.ReferrerID,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified, login, password_hash
		FROM users
		WHERE phone = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, phone)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
		&user.Email,
		&user.Phone,
		&user.IsEmailVerified,
		&user.IsPhoneVerified,
		&user.Login,
		&user.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified, login, password_hash
		FROM users
		WHERE login = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, login)

	var user User
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
		&user.Email,
		&user.Phone,
		&user.IsEmailVerified,
		&user.IsPhoneVerified,
		&user.Login,
		&user.PasswordHash,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresRepository) SetEmailVerified(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET is_email_verified = true
		WHERE id = $1
	`
	_, err := r.DB.Pool.Exec(ctx, query, userID)
	return err
}

func (r *PostgresRepository) SetPhoneVerified(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET is_phone_verified = true
		WHERE id = $1
	`
	_, err := r.DB.Pool.Exec(ctx, query, userID)
	return err
}

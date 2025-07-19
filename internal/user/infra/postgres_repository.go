package user

import (
	"context"
	"log"

	"github.com/Vovarama1992/emelya-go/internal/db"
	model "github.com/Vovarama1992/emelya-go/internal/user/model"
)

type UserRepository struct {
	DB *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (
			first_name, last_name, patronymic, email, phone,
			is_email_verified, is_phone_verified, login, password_hash, referrer_id, role
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
		user.Role,
	).Scan(&user.ID)

	return err
}

func (r *UserRepository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified,
		       is_phone_verified, login, password_hash, referrer_id, card_number, balance, role
		FROM users
		WHERE id = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, userID)

	var user model.User
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
		&user.CardNumber,
		&user.Balance,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified,
		       login, password_hash, referrer_id, card_number, role
		FROM users
		WHERE phone = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, phone)

	var user model.User
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
		&user.CardNumber,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified,
		       login, password_hash, card_number, role
		FROM users
		WHERE login = $1
	`
	row := r.DB.Pool.QueryRow(ctx, query, login)

	var user model.User
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
		&user.CardNumber,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) SetEmailVerified(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET is_email_verified = true
		WHERE id = $1
	`
	_, err := r.DB.Pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepository) SetPhoneVerified(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET is_phone_verified = true
		WHERE id = $1
	`
	_, err := r.DB.Pool.Exec(ctx, query, userID)
	return err
}

func (r *UserRepository) UpdateBalance(ctx context.Context, userID int64, balance float64) error {
	query := `UPDATE users SET balance = $1 WHERE id = $2`
	_, err := r.DB.Pool.Exec(ctx, query, balance, userID)
	return err
}

func (r *UserRepository) UpdateCardNumber(ctx context.Context, userID int64, cardNumber string) error {
	query := `UPDATE users SET card_number = $1 WHERE id = $2`
	_, err := r.DB.Pool.Exec(ctx, query, cardNumber, userID)
	return err
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, email, phone, is_email_verified,
		       is_phone_verified, login, password_hash, referrer_id, card_number, balance, role
		FROM users
	`
	rows, err := r.DB.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("Ошибка запроса GetAllUsers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
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
			&user.CardNumber,
			&user.Balance,
			&user.Role,
		)
		if err != nil {
			log.Printf("Ошибка сканирования GetAllUsers: %v", err)
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

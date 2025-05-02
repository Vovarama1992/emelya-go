package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New() (*DB, error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return nil, ErrMissingDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Подключение к базе данных успешно")
	return &DB{Pool: pool}, nil
}

var ErrMissingDatabaseURL = &DatabaseError{"DATABASE_URL не найден в окружении"}

type DatabaseError struct {
	Message string
}

func (e *DatabaseError) Error() string {
	return e.Message
}

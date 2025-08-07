package storage

import (
	"database/sql"
	"fmt"
	"order/config"

	_ "github.com/lib/pq"
)

// NewDatabaseConnection устанавливает соединение с базой данных
func NewDatabaseConnection(cfg *config.Config) (*sql.DB, Store, error) {
	switch cfg.DB.Type {
	case config.Postgres:
		return NewPostgresRepository(cfg)
	default:
		return nil, nil, fmt.Errorf("unsupported database type: %s", cfg.DB.Type)
	}
}

// NewPostgresRepository создает подключение к Postgres
func NewPostgresRepository(cfg *config.Config) (*sql.DB,Store, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.Mode,
	)
	db, err := sql.Open(cfg.DB.SType, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, &Storage{db: db}, nil
}

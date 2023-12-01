package database

import (
	"database/sql"
	"time"
)

type Service interface {
	TestConnection() error
	Close() error
	GetDB() (*sql.DB, error)
}

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

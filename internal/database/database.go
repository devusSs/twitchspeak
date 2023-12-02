package database

import (
	"database/sql"
	"time"
)

type Service interface {
	TestConnection() error
	Close() error
	GetDB() (*sql.DB, error)

	Migrate() error

	AddUser(user *User) (*User, error)
	GetUserByTwitchID(twichID string) (*User, error)

	// TODO: functions to integrate teamspeak details
}

// TODO: add more fields like the teamspeak details
type User struct {
	ID        uint      `gorm:"primarykey" json:"-"`
	CreatedAt time.Time `                  json:"connected_since"`
	UpdatedAt time.Time `                  json:"-"`

	TeamSpeakUID string `gorm:"uniqueIndex" json:"teamspeak_uid"`
	TwitchID     string `gorm:"uniqueIndex" json:"twitch_id"`
}

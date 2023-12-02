package psql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/devusSs/twitchspeak/internal/database"
	"github.com/devusSs/twitchspeak/pkg/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type psql struct {
	db *gorm.DB
}

func (p *psql) TestConnection() error {
	db, err := p.db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func (p *psql) Close() error {
	db, err := p.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (p *psql) GetDB() (*sql.DB, error) {
	db, err := p.db.DB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (p *psql) Migrate() error {
	return p.db.AutoMigrate(&database.User{})
}

func (p *psql) AddUser(user *database.User) (*database.User, error) {
	err := p.db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *psql) GetUserByTwitchID(twitchID string) (*database.User, error) {
	var user database.User
	err := p.db.Where("twitch_id = ?", twitchID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Config for the database service
type Config struct {
	Host     string
	Port     uint
	User     string
	Password string
	Database string
	Console  bool
	Debug    bool
}

// NewService returns a new database service
func NewService(cfg Config) (database.Service, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
	fLogger := log.NewLogger(
		log.WithOwnLogFile("database.log"),
		log.WithName("data"),
		log.WithConsole(cfg.Console),
		log.WithDebug(cfg.Debug),
	)
	l := logger.New(fLogger, logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Silent,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: l,
	})
	if err != nil {
		return nil, err
	}
	return &psql{
		db: db,
	}, nil
}

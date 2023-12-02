package teamspeak

import (
	"context"
	"fmt"
	"sync"

	"github.com/multiplay/go-ts3"

	"github.com/devusSs/twitchspeak/internal/database"
	"github.com/devusSs/twitchspeak/pkg/log"
)

// BotConfig is the configuration for the bot
type BotConfig struct {
	Host      string
	Queryport uint
	Port      uint
	// This is NOT the nickname but server query login name
	Username string
	Password string
	Nickname string
	// API specific login url for connecting Twitch
	LoginBaseURL string
	DB           database.Service
	Console      bool
	Debug        bool
}

// Bot is the bot
type Bot struct {
	host      string
	queryport uint
	port      uint
	username  string
	password  string
	nickname  string

	loginBaseURL string

	logger *log.Logger
	db     database.Service
	client *ts3.Client
}

// EstablishConn establishes a connection to the TeamSpeak server
// and sets up the client's port and nickname
func (b *Bot) EstablishConn() error {
	client, err := ts3.NewClient(fmt.Sprintf("%s:%d", b.host, b.queryport))
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	b.client = client

	b.logger.Debug("created new client connection to %s:%d", b.host, b.queryport)

	if err := b.client.Login(b.username, b.password); err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	b.logger.Debug("logged in")

	if err := b.client.UsePort(int(b.port)); err != nil {
		return fmt.Errorf("using port: %w", err)
	}

	b.logger.Debug("using port: %d", b.port)

	if err := b.client.SetNick(b.nickname); err != nil {
		return fmt.Errorf("setting nickname: %w", err)
	}

	b.logger.Debug("set nickname: %s", b.nickname)

	b.logger.Info("Bot connected and initialized")

	return nil
}

// RegisterEvents registers events for the bot
func (b *Bot) RegisterEvents() error {
	events := []ts3.NotifyCategory{
		ts3.ServerEvents,
		ts3.ChannelEvents,
		ts3.TextServerEvents,
		ts3.TextChannelEvents,
		ts3.TextPrivateEvents,
		ts3.TokenUsedEvents,
	}

	for _, event := range events {
		if err := b.client.Register(event); err != nil {
			return fmt.Errorf("registering event: %w", err)
		}
	}

	b.logger.Info("Registered events: %v", events)

	return nil
}

// HandleEvents handles events from the TeamSpeak server
//
// Blocks until context is canceled
func (b *Bot) HandleEvents(ctx context.Context, wg *sync.WaitGroup) {
	b.logger.Info("Setup event handler")
	for {
		select {
		case <-ctx.Done():
			b.logger.Debug("context done, exiting event handler")
			wg.Done()
			return
		// TODO: setup event handlers
		case event := <-b.client.Notifications():
			b.logger.Debug("Event: %v", event)
		}
	}
}

// NewBot creates a new bot but does not connect it
func NewBot(cfg BotConfig) *Bot {
	logger := log.NewLogger(
		log.WithOwnLogFile("teamspeak.log"),
		log.WithName("ts3"),
		log.WithConsole(cfg.Console),
		log.WithDebug(cfg.Debug),
	)

	bot := &Bot{
		host:      cfg.Host,
		queryport: cfg.Queryport,
		port:      cfg.Port,
		username:  cfg.Username,
		password:  cfg.Password,
		nickname:  cfg.Nickname,

		loginBaseURL: cfg.LoginBaseURL,

		logger: logger,
		db:     cfg.DB,
	}

	return bot
}

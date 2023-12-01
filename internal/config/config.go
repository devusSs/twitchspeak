package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config is the struct that holds the configuration for the application.
type Config struct {
	APIPort     uint   `env:"API_PORT"     envDefault:"8080"                  print:"true"`
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:5173" print:"true"`
	BackendURL  string `env:"BACKEND_URL"  envDefault:"http://localhost:8080" print:"true"`
	SecretKey   string `env:"SECRET_KEY"                                      print:"false"`

	TwitchClientID     string `env:"TWITCH_CLIENT_ID"     print:"false"`
	TwitchClientSecret string `env:"TWITCH_CLIENT_SECRET" print:"false"`
	TwitchRedirectURI  string `env:"TWITCH_REDIRECT_URI"  print:"true"`

	PostgresHost     string `env:"POSTGRES_HOST"     envDefault:"localhost" print:"true"`
	PostgresPort     uint   `env:"POSTGRES_PORT"     envDefault:"5432"      print:"true"`
	PostgresUser     string `env:"POSTGRES_USER"                            print:"false"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"                        print:"false"`
	PostgresDB       string `env:"POSTGRES_DB"                              print:"false"`

	RedisHost     string `env:"REDIS_HOST"     envDefault:"localhost" print:"true"`
	RedisPort     uint   `env:"REDIS_PORT"     envDefault:"6379"      print:"true"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""          print:"false"`
	RedisDB       uint   `env:"REDIS_DB"       envDefault:"0"         print:"true"`

	TeamspeakHost      string `env:"TEAMSPEAK_HOST"       envDefault:"localhost" print:"true"`
	TeamspeakQueryPort uint   `env:"TEAMSPEAK_QUERY_PORT" envDefault:"10011"     print:"true"`
	TeamspeakPort      uint   `env:"TEAMSPEAK_PORT"       envDefault:"9987"      print:"true"`
	TeamspeakUser      string `env:"TEAMSPEAK_USER"                              print:"false"`
	TeamspeakPassword  string `env:"TEAMSPEAK_PASSWORD"                          print:"false"`
	TeamspeakNickname  string `env:"TEAMSPEAK_NICKNAME"                          print:"true"`
}

// String returns the string representation of the config struct.
func (c *Config) String() string {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	m := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("print")

		// We use the env tag since env is the default
		if tag == "true" {
			m[t.Field(i).Tag.Get("env")] = field.Interface()
		}
	}

	content, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("error marshalling config: %v", err)
	}

	return string(content)
}

// Load determines the file type of provided config and loads config accordingly.
func Load(path string) (*Config, error) {
	switch filepath.Ext(path) {
	case ".env":
		return loadEnv(path)
	default:
		return nil, fmt.Errorf("unknown config file type: %s", filepath.Ext(path))
	}
}

func loadEnv(envFile ...string) (*Config, error) {
	fileProvided := len(envFile) > 0
	if fileProvided {
		if err := godotenv.Load(envFile[0]); err != nil {
			return nil, fmt.Errorf("loading env file: %w", err)
		}
	}
	var cfg Config
	if err := env.ParseWithOptions(&cfg, envOptions); err != nil {
		return nil, fmt.Errorf("parsing env: %w", err)
	}
	return &cfg, nil
}

var (
	envOptions = env.Options{
		Prefix:          "TWITCHSPEAK_",
		RequiredIfNoDef: true,
	}
)

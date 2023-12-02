package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/devusSs/twitchspeak/internal/auth/twitch"
	"github.com/devusSs/twitchspeak/internal/config"
	"github.com/devusSs/twitchspeak/internal/database/psql"
	"github.com/devusSs/twitchspeak/internal/database/redis"
	"github.com/devusSs/twitchspeak/internal/server"
	"github.com/devusSs/twitchspeak/internal/updater"
	"github.com/devusSs/twitchspeak/pkg/log"
	"github.com/devusSs/twitchspeak/pkg/system"
)

func main() {
	helpFlag := flag.Bool("help", false, "Prints help information and exits")
	versionFlag := flag.Bool("version", false, "Prints version information and exits")
	noUpdateFlag := flag.Bool("no-update", false, "Disables automatic update checks")
	consoleFlag := flag.Bool("console", false, "Enables log output to console")
	debugFlag := flag.Bool("debug", false, "Enables debug mode (verbose logging, also to console)")
	logsDirFlag := flag.StringP("logs", "l", "logs", "Directory to store logs in")
	configFileFlag := flag.StringP("config", "c", "", "Path to config file")
	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		printVersion()
		os.Exit(0)
	}

	if err := checkOSAndArch(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	updateAvailable := make(chan bool, 1)

	if !*noUpdateFlag {
		if err := updater.CheckForUpdatesAndApply(buildVersion); err != nil {
			fmt.Println("Error checking for updates:", err)
			os.Exit(1)
		}
		wg.Add(1)
		go updater.PeriodicUpdateCheck(ctx, buildVersion, updateAvailable, wg)
	}

	log.SetDefaultLogsDirectory(*logsDirFlag)
	log.SetDefaultLogFileName("twitchspeak.log")
	logger := log.NewLogger(
		log.WithName("main"),
		log.WithConsole(*consoleFlag),
		log.WithDebug(*debugFlag),
	)

	cfg, err := config.Load(*configFileFlag)
	if err != nil {
		logger.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	logger.Debug("loaded config: %v", cfg)
	logger.Info("Config loaded successfully")

	svc, err := psql.NewService(psql.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		Database: cfg.PostgresDB,
		Console:  *consoleFlag,
		Debug:    *debugFlag,
	})
	if err != nil {
		logger.Error("Error initializing database service: %v", err)
		os.Exit(1)
	}

	if err := svc.TestConnection(); err != nil {
		logger.Error("Error testing database connection: %v", err)
		os.Exit(1)
	}

	if err := svc.Migrate(); err != nil {
		logger.Error("Error migrating database: %v", err)
		os.Exit(1)
	}

	err = redis.Init(redis.Config{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		logger.Error("Error initializing redis: %v", err)
		os.Exit(1)
	}

	err = twitch.Init(twitch.Config{
		ClientID:     cfg.TwitchClientID,
		ClientSecret: cfg.TwitchClientSecret,
		RedirectURI:  cfg.TwitchRedirectURI,
		Port:         cfg.APIPort,
		FrontendURL:  cfg.FrontendURL,
		Svc:          svc,
	})
	if err != nil {
		logger.Error("Error initializing twitch oauth: %v", err)
		os.Exit(1)
	}

	s := server.NewServer(server.Config{
		Port:        cfg.APIPort,
		BackendURL:  cfg.BackendURL,
		FrontendURL: cfg.FrontendURL,
		Console:     *consoleFlag,
		Debug:       *debugFlag,
	})

	if err := s.ApplyMiddlewares(svc, cfg.SecretKey); err != nil {
		logger.Error("Error applying middlewares: %v", err)
		os.Exit(1)
	}

	if err := s.SetupRoutes(cfg.TwitchRedirectURI, svc); err != nil {
		logger.Error("Error setting up routes: %v", err)
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT)

	errChan := make(chan error)

	wg.Add(1)
	go s.Start(ctx, errChan, wg)

	for {
		select {
		case sig := <-stop:
			fmt.Println()
			logger.Info("Received signal '%s', stopping...", sig.String())
			cancel()
			wg.Wait()
			close(stop)
			if err := svc.Close(); err != nil {
				logger.Error("Error closing database connection: %v", err)
				os.Exit(1)
			}
			logger.Debug("Database connection closed")
			logger.Info("Shutdown complete")
			os.Exit(0)
		case err := <-errChan:
			if err == server.ErrorCritical {
				logger.Error("Critical error: %v", err)
				cancel()
				wg.Wait()
				close(stop)
				if err := svc.Close(); err != nil {
					logger.Error("Error closing database connection: %v", err)
					os.Exit(1)
				}
				logger.Debug("Database connection closed")
				os.Exit(1)
			}
			logger.Error("Error: %v", err)
		case <-updateAvailable:
			logger.Info("New update available, please restart the app")
		}
	}
}

const appMessage = `twitchspeak - Twitch integration for TeamSpeak 3`

var (
	buildVersion   string
	buildDate      string
	buildGitCommit string
)

func init() {
	if buildVersion == "" {
		buildVersion = "dev"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}
	if buildGitCommit == "" {
		buildGitCommit = "unknown"
	}
}

func printHelp() {
	fmt.Println(appMessage)
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  twitchspeak [FLAGS]")
	fmt.Println()
	fmt.Println("FLAGS:")
	flag.PrintDefaults()
}

func printVersion() {
	fmt.Println(appMessage)
	fmt.Println()
	fmt.Printf("Build version:\t\t%s\n", buildVersion)
	fmt.Printf("Build date:\t\t%s\n", buildDate)
	fmt.Printf("Build Git commit:\t%s\n", buildGitCommit)
	fmt.Println()
	fmt.Printf("Build Go version:\t%s\n", runtime.Version())
	fmt.Printf("Build Go os:\t\t%s\n", runtime.GOOS)
	fmt.Printf("Build Go arch:\t\t%s\n", runtime.GOARCH)
}

var (
	supportedOS   = []string{"Linux", "Windows", "macOS"}
	supportedArch = []string{"amd64", "x86_64", "arm64", "aarch64"}
)

func checkOSAndArch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	osV, err := system.GetOS(ctx)
	if err != nil {
		return fmt.Errorf("error getting OS: %v", err)
	}

	archV, err := system.GetArch(ctx)
	if err != nil {
		return fmt.Errorf("error getting arch: %v", err)
	}

	if !slices.Contains(supportedOS, osV) {
		return fmt.Errorf("unsupported OS: %s", osV)
	}

	if !slices.Contains(supportedArch, archV) {
		return fmt.Errorf("unsupported arch: %s", archV)
	}

	if osV == "Windows" && archV != "amd64" && archV != "x86_64" {
		return fmt.Errorf("unsupported arch for Windows: %s", archV)
	}

	return nil
}

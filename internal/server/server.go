package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"

	"github.com/devusSs/twitchspeak/internal/auth/twitch"
	"github.com/devusSs/twitchspeak/internal/database"
	"github.com/devusSs/twitchspeak/internal/database/redis"
	"github.com/devusSs/twitchspeak/internal/server/responses"
	"github.com/devusSs/twitchspeak/internal/server/routes"
	"github.com/devusSs/twitchspeak/pkg/log"
)

// Custom errors
var (
	ErrorCritical = fmt.Errorf("critical error")
)

// Config for the http server
type Config struct {
	Port        uint
	BackendURL  string
	FrontendURL string
	Console     bool
	Debug       bool
}

// Server is the main struct for the http server
// wrapped around Gin
type Server struct {
	port uint
	// Host (host:port) of the backend server (for sessions purposes)
	backendURL string
	// Host (host:port) of the frontend server (for cors purposes)
	frontendURl string

	logger *log.Logger
	engine *gin.Engine
}

// Applies middlewares to the gin engine
// like recovery, custom logging, cors, rate limiting and sessions
func (s *Server) ApplyMiddlewares(svc database.Service, secretKey string) error {
	s.engine.Use(gin.Recovery())
	s.engine.Use(s.customLogger())
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{s.frontendURl},
		AllowMethods:     []string{http.MethodGet, http.MethodPatch},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	rStore := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: redis.GetClient(),
		Rate:        time.Second,
		Limit:       3,
	})

	keyFunc := func(c *gin.Context) string {
		return c.ClientIP()
	}

	errorHandler := func(c *gin.Context, info ratelimit.Info) {
		c.JSON(http.StatusTooManyRequests, responses.Error{
			Code:      http.StatusTooManyRequests,
			ErrorCode: "too_many_requests",
			ErrorMessage: fmt.Sprintf(
				"rate limit hit, wait %v",
				time.Until(info.ResetTime),
			),
		})
	}

	mw := ratelimit.RateLimiter(rStore, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})

	s.engine.Use(mw)

	db, err := svc.GetDB()
	if err != nil {
		return fmt.Errorf("could not get database connection: %v", err)
	}

	store, err := postgres.NewStore(db, []byte(secretKey))
	if err != nil {
		return fmt.Errorf("could not create postgres store: %v", err)
	}

	s.engine.Use(sessions.Sessions("twitchspeak", store))

	s.logger.Info("Applied middlewares successfully")

	return nil
}

// SetupRoutes sets up the routes for the gin engine
func (s *Server) SetupRoutes(twitchRedirectURI string, svc database.Service) error {
	if twitchRedirectURI == "" {
		return fmt.Errorf("twitch redirect uri is empty")
	}

	routes.Svc = svc

	u, err := url.Parse(twitchRedirectURI)
	if err != nil {
		return fmt.Errorf("invalid twitch redirect uri: %v", err)
	}

	path := u.Path
	if strings.Contains(path, "auth/") {
		if strings.Count(path, "auth/") > 0 {
			path = strings.Replace(path, "auth/", "", 1)
		}
	}

	s.engine.NoRoute(routes.NoRoute)
	s.engine.NoMethod(routes.NoMethod)

	base := s.engine.Group("/")
	{
		base.GET("/", routes.HomeRoute)

		auth := base.Group("/auth")
		{
			auth.GET("/twitch/login", twitch.HandleLoginRoute)
			auth.GET(path, twitch.HandleRedirectRoute)
			auth.GET("/logout", routes.LogoutRoute)
		}

		users := base.Group("/users")
		{
			users.GET("/me", routes.GetMeRoute)
		}
	}

	s.logger.Info("Setup routes properly")

	return nil
}

// Start starts the server and listens for incoming requests
//
// Blocks until the context is canceled or a critical error occurs
func (s *Server) Start(ctx context.Context, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", s.port),
		Handler: s.engine,
	}

	go func() {
		s.logger.Info("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("%v: %w", ErrorCritical, err)
			close(errChan)
			return
		}
	}()

	<-ctx.Done()

	s.logger.Debug("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		errChan <- fmt.Errorf("%v: %w", ErrorCritical, err)
		close(errChan)
		return
	}

	close(errChan)
	s.logger.Debug("Server shutdown complete")
}

func (s *Server) customLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		elapsed := time.Since(start)
		format := fmt.Sprintf(
			"[%s] %s %s %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			elapsed,
		)
		_, _ = s.logger.GetWriter().
			Write([]byte(format))
	}
}

// NewServer creates a new server instance
func NewServer(cfg Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	logger := log.NewLogger(
		log.WithOwnLogFile("server.log"),
		log.WithName("http"),
		log.WithConsole(cfg.Console),
		log.WithDebug(cfg.Debug),
	)

	engine := gin.New()
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = false
	engine.HandleMethodNotAllowed = true
	engine.ForwardedByClientIP = true
	engine.UseRawPath = false
	engine.UnescapePathValues = true

	s := &Server{
		port:        cfg.Port,
		backendURL:  cfg.BackendURL,
		frontendURl: cfg.FrontendURL,
		logger:      logger,
		engine:      engine,
	}

	s.logger.Info("Server initialized successfully")

	return s
}

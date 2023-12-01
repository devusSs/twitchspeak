package twitch

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	"github.com/devusSs/twitchspeak/internal/database"
	"github.com/devusSs/twitchspeak/internal/httplib"
	"github.com/devusSs/twitchspeak/internal/server/responses"
)

// Config for oauth2 authorization process
// and hosting auth routes
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string

	Port        uint
	FrontendURL string

	Svc database.Service
}

// Initializes our oauth2 config
func Init(cfg Config) error {
	if cfg.ClientID == "" {
		return fmt.Errorf("twitch client id is empty")
	}

	if cfg.ClientSecret == "" {
		return fmt.Errorf("twitch client secret is empty")
	}

	if cfg.RedirectURI == "" {
		return fmt.Errorf("twitch redirect uri is empty")
	}

	if cfg.Port == 0 {
		return fmt.Errorf("twitch: port is empty")
	}

	redirectU, err := url.Parse(cfg.RedirectURI)
	if err != nil {
		return fmt.Errorf("twitch: invalid redirect uri: %v", err)
	}

	if fmt.Sprintf(
		"%s:%s",
		redirectU.Hostname(),
		redirectU.Port(),
	) != fmt.Sprintf(
		"localhost:%d",
		cfg.Port,
	) {
		return fmt.Errorf("twitch: redirect uri does not match host url")
	}

	if cfg.FrontendURL == "" {
		return fmt.Errorf("twitch: frontend url is empty")
	}

	frontendURL = cfg.FrontendURL
	svc = cfg.Svc

	oauthConfig = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Endpoint:     endpoints.Twitch,
		Scopes:       []string{"openid"},
	}

	return nil
}

// HandleLoginRoute handles the login route
func HandleLoginRoute(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("twitch_id")

	if id != nil {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL)
		return
	}

	if oauthConfig == nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}
	requests.set(c.ClientIP(), request{
		nonce: generateRandomString(nonceLength),
		state: generateRandomString(stateLength),
	})
	url := oauthConfig.AuthCodeURL(
		requests.get(c.ClientIP()).state,
		oauth2.SetAuthURLParam("nonce", requests.get(c.ClientIP()).nonce),
	)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleRedirectRoute handles the redirect route
func HandleRedirectRoute(c *gin.Context) {
	session := sessions.Default(c)
	id := session.Get("twitch_id")

	if id != nil {
		c.Redirect(http.StatusTemporaryRedirect, frontendURL)
		return
	}

	if oauthConfig == nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	qState := c.Query("state")
	qCode := c.Query("code")

	if qState != requests.get(c.ClientIP()).state {
		resp := responses.Error{
			Code:         http.StatusBadRequest,
			ErrorCode:    "invalid_state",
			ErrorMessage: "State does not match required",
		}
		c.JSON(resp.Code, resp)
		return
	}

	token, err := oauthConfig.Exchange(c, qCode)
	if err != nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	nonce := token.Extra("nonce")
	if nonce != requests.get(c.ClientIP()).nonce {
		resp := responses.Error{
			Code:         http.StatusBadRequest,
			ErrorCode:    "invalid_nonce",
			ErrorMessage: "Nonce does not match required",
		}
		c.JSON(resp.Code, resp)
		return
	}

	var claims claimsResponse
	err = httplib.AuthorizedGet(
		http.MethodGet,
		userInfoEndpoint,
		fmt.Sprintf("%v", token.Extra("access_token")),
		&claims,
	)
	if err != nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	twitchID := token.Extra("id_token")

	u, err := url.Parse(c.Request.RequestURI)
	if err != nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	session.Options(sessions.Options{
		Path: "/",
		// Might be dropped on dev since host:port is not a valid domain
		Domain: strings.Replace(
			strings.Replace(u.Host, "https://", "", 1),
			"http://",
			"",
			1,
		),
		MaxAge:   86400 * 7,
		Secure:   strings.Contains(u.String(), "https://"),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	session.Set("twitch_id", twitchID)
	if err := session.Save(); err != nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	// TODO: add stuff to database using svc
	fmt.Println(claims)
	fmt.Println(claims.Sub)

	// TODO: remove this, just for linting purposes
	if err := svc.TestConnection(); err != nil {
		fmt.Println("RANDOM ERROR LELELELELELE")
	}

	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

const (
	stateLength = 16
	nonceLength = 32
)

var (
	frontendURL string           = ""
	svc         database.Service = nil
	oauthConfig *oauth2.Config   = nil
	// Maps request ip to request (nonce and state)
	requests *safeMap = &safeMap{mu: sync.Mutex{}, data: make(map[string]request)}

	userInfoEndpoint = "https://id.twitch.tv/oauth2/userinfo"
)

type safeMap struct {
	mu   sync.Mutex
	data map[string]request
}

func (s *safeMap) get(key string) request {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.data[key]
	if !ok {
		return request{}
	}
	return val
}

func (s *safeMap) set(key string, value request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

type request struct {
	nonce string
	state string
}

func generateRandomString(length int) string {
	randomBytes := make([]byte, length)
	_, _ = rand.Read(randomBytes)
	randomString := base64.URLEncoding.EncodeToString(randomBytes)
	return randomString[:length]
}

type claimsResponse struct {
	Aud           string    `json:"aud"`
	Exp           int       `json:"exp"`
	Iat           int       `json:"iat"`
	Iss           string    `json:"iss"`
	Sub           string    `json:"sub"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Picture       string    `json:"picture"`
	UpdatedAt     time.Time `json:"updated_at"`
}

package routes

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/devusSs/twitchspeak/internal/server/responses"
)

// NoRoute handles requests with invalid routes
func NoRoute(c *gin.Context) {
	resp := responses.Error{
		Code:         http.StatusNotFound,
		ErrorCode:    "not_found",
		ErrorMessage: "The requested resource was not found",
	}
	c.JSON(resp.Code, resp)
}

// NoMethod handles requests with invalid methods
func NoMethod(c *gin.Context) {
	resp := responses.Error{
		Code:         http.StatusMethodNotAllowed,
		ErrorCode:    "method_not_allowed",
		ErrorMessage: "The requested method is not allowed",
	}
	c.JSON(resp.Code, resp)
}

// HomeRoute handles requests to the home route
func HomeRoute(c *gin.Context) {
	err := c.Query("error")
	errCode := c.Query("error_code")

	if err != "" {
		resp := responses.Error{
			Code:         http.StatusBadRequest,
			ErrorCode:    errCode,
			ErrorMessage: err,
		}
		c.JSON(resp.Code, resp)
		return
	}

	session := sessions.Default(c)
	id := session.Get("twitch_id")

	data := struct {
		LoggedIn bool   `json:"logged_in"`
		Message  string `json:"message"`
	}{id != nil, "Twitchspeak API"}

	resp := responses.Success{
		Code: http.StatusOK,
		Data: data,
	}
	c.JSON(resp.Code, resp)
}

// LogoutRoute handles requests to the logout route
func LogoutRoute(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()

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
		MaxAge:   -1,
		Secure:   strings.Contains(u.String(), "https://"),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	if err := session.Save(); err != nil {
		resp := responses.Error{
			Code:         http.StatusInternalServerError,
			ErrorCode:    responses.CodeInternalError,
			ErrorMessage: responses.MessageInternalError,
		}
		c.JSON(resp.Code, resp)
		return
	}

	resp := responses.Success{
		Code: http.StatusOK,
		Data: "Successfully logged out",
	}
	c.JSON(resp.Code, resp)
}

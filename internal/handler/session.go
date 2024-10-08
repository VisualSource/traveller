package handler

import (
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"visualsource/traveller/internal/model"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/lucsky/cuid"
)

const SESSION_NAME = "session"

func (h *Handler) View(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Render(http.StatusUnauthorized, "session_error.html", map[string]interface{}{
			"message": "Invalid user session",
			"code":    http.StatusUnauthorized,
		})
	}

	user := sess.Values[Session_ID].(string)

	// Get the session id from HX-Current-URL header as this is being load via htmx load trigger
	currentURL := c.Request().Header.Get("HX-Current-URL")

	url, err := url.Parse(currentURL)
	if err != nil {
		return c.Render(http.StatusBadRequest, "session_error.html", map[string]interface{}{
			"message": "Unable to determin session (Missing Header 'HX-Current-URL')",
			"code":    http.StatusBadRequest,
		})
	}

	path := strings.Split(url.Path, "/")
	if len(path) != 3 {
		return c.Render(http.StatusBadRequest, "session_error.html", map[string]interface{}{
			"message": "Unable to get session (1)",
			"code":    http.StatusBadRequest,
		})
	}

	if path[1] != "session" {
		return c.Render(http.StatusBadRequest, "session_error.html", map[string]interface{}{
			"message": "Unable to get session (2)",
			"code":    http.StatusBadRequest,
		})
	}

	id := path[2]

	var session model.Session
	log.Printf("Session id is: %s", id)
	err = session.GetSession(h.db, id)
	if err != nil {
		return c.Render(http.StatusNotFound, "session_error.html", map[string]interface{}{
			"message": "No session found",
			"code":    http.StatusNotFound,
		})
	}

	if session.Admin == user {
		return c.Render(http.StatusOK, "session_cmd.html", map[string]interface{}{
			"user": user,
		})
	}

	if !slices.Contains(session.Players, user) {
		return c.HTML(http.StatusUnauthorized, `<span style="color:red;">Unauthorized</span>`)
	}

	if flashes := sess.Flashes(); len(flashes) > 0 {
		log.Println(flashes)

		return c.Render(http.StatusOK, "session_create_player.html", map[string]interface{}{
			"user": user,
		})
	}

	return c.Render(http.StatusOK, "session_player.html", map[string]interface{}{
		"user": user,
	})
}

type sessionForm struct {
	SessionName string `form:"session_name"`
}

func (h *Handler) CreateSession(c echo.Context) error {
	var formData sessionForm
	err := c.Bind(&formData)
	if err != nil {
		return c.String(http.StatusBadRequest, "missing form fields")
	}

	sess, err := session.Get(SESSION_NAME, c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	sessionId := cuid.Slug()

	stmt, err := h.db.Prepare("INSERT INTO session ('id','name','admin','players') VALUES (?,?,?,?)")

	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, "Server Error")
	}

	_, err = stmt.Exec(sessionId, formData.SessionName, sess.Values[Session_ID].(string), "[]")
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, "Server Error")
	}

	c.Response().Header().Add("HX-Redirect", strings.Join([]string{"/session/in/", sessionId}, ""))

	return c.String(http.StatusCreated, "Created")
}

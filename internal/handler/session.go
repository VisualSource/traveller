package handler

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) View(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	user := sess.Values["sub"]

	stmt, err := h.db.Prepare("SELECT admin FROM session WHERE id = ? LIMIT 1;")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to prepare query")
	}

	var admin string
	err = stmt.QueryRow(c.Param("id")).Scan(&admin)

	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid session id")
	}

	if admin == user {
		return c.Render(http.StatusOK, "session/cmd.html", map[string]interface{}{
			"user": user,
		})
	}

	return c.Render(http.StatusOK, "session/player.html", map[string]interface{}{
		"user": user,
	})
}

func (h *Handler) CreateSession(c echo.Context) error {
	return c.String(http.StatusCreated, "Created")
}

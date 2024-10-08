package traveller

import (
	"net/http"
	handler "visualsource/traveller/internal/handler"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func sessionRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return err
		}

		if sess.Values["sub"] == nil {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		return next(c)
	}
}

func RegisterSessionPages(e *echo.Echo, h *handler.Handler) {
	g := e.Group("/session")

	g.Use(sessionRequired)

	// html pages
	g.File("/select", "web/app//session_select.html")
	g.File("/create", "web/app/session_create.html")

	g.File("/:id", "web/app/session.html")

	g.GET("/view", h.View).Name = "session-view"
	// curd end points
	g.POST("/join", h.UserJoinSession).Name = "session-join"
	g.GET("/joined", h.UserJoinedSessions).Name = "user-sessions"

	g.POST("/cmd/create", h.CreateSession).Name = "create-session"

	g.GET("/:sessionId/ws", h.WS).Name = "ws"
}

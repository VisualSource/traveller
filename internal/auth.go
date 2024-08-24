package traveller

import (
	handler "visualsource/traveller/internal/handler"

	"github.com/labstack/echo/v4"
)

type UserLogin struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func RegisterAuthPages(e *echo.Echo, h *handler.Handler) {

	// Pages
	e.File("/account", "public/account.html")

	// End Points

	e.GET("/account/logout", h.UserLogout)

	e.POST("/account/login", h.UserLogin)
	e.POST("/account/signup", h.UserSignup)
}

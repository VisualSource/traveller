package handler

import (
	"visualsource/traveller/internal/socket"

	"github.com/labstack/echo/v4"
)

func (h *Handler) WS(c echo.Context) error {
	return socket.ServeWs(h.hub, c)
}

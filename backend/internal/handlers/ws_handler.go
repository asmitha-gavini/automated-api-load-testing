package handlers

import (
	"backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

// WSHandler encapsulates WebSocket upgrade routes.
type WSHandler struct {
	hub *websocket.Hub
}

// NewWSHandler creates a new WSHandler instance.
func NewWSHandler(hub *websocket.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleTestWS upgrades HTTP connections to stream metrics for a specific test ID.
func (h *WSHandler) HandleTestWS(c *gin.Context) {
	testID := c.Param("testId")
	if testID == "" {
		testID = c.Param("id")
	}
	websocket.ServeWS(h.hub, c.Writer, c.Request, testID)
}

// HandleGeneralWS upgrades HTTP connections to stream all live metrics without ID filtering.
func (h *WSHandler) HandleGeneralWS(c *gin.Context) {
	websocket.ServeWS(h.hub, c.Writer, c.Request, "*")
}

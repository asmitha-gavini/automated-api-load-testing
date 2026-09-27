package handlers

import (
	"net/http"
	"time"

	"backend/internal/database"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *database.DB
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck responds with current service status and database connectivity.
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	dbStatus := "connected"
	if h.db != nil {
		if err := h.db.Ping(c.Request.Context()); err != nil {
			dbStatus = "disconnected"
		}
	} else {
		dbStatus = "uninitialized"
	}

	status := "ok"
	httpCode := http.StatusOK
	if dbStatus != "connected" {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
	}

	c.JSON(httpCode, gin.H{
		"status":    status,
		"service":   "automated-api-load-testing-backend",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"database":  dbStatus,
	})
}

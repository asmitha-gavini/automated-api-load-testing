package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/engine"
	"backend/internal/handlers"
	"backend/internal/metrics"
	"backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting Automated API Load Testing Backend [Env: %s, Port: %s]", cfg.Environment, cfg.Port)

	// 2. Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 3. Initialize SQLite database & migrations
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()
	log.Printf("SQLite database initialized successfully at %s", cfg.DBPath)

	// 4. Initialize Prometheus Telemetry & WebSocket Hub
	promMetrics := metrics.NewPrometheusMetrics()
	wsHub := websocket.NewHub()
	go wsHub.Run()
	defer wsHub.Close()

	// 5. Initialize Core Load Engine with WebSocket and Prometheus integration
	loadEngine := engine.NewEngine(db, wsHub, promMetrics)

	// 6. Setup Gin engine and middlewares
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg.AllowedOrigins))

	// 7. Register Handlers
	healthHandler := handlers.NewHealthHandler(db)
	testHandler := handlers.NewTestHandler(loadEngine, db)
	wsHandler := handlers.NewWSHandler(wsHub)

	// Health and Prometheus Observability endpoints
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/metrics", gin.WrapH(promMetrics.Handler()))

	// WebSocket endpoints
	router.GET("/ws/tests/:testId", wsHandler.HandleTestWS)
	router.GET("/ws/tests", wsHandler.HandleGeneralWS)
	router.GET("/ws/metrics", wsHandler.HandleGeneralWS)

	// REST API v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.HealthCheck)
		v1.GET("/metrics", gin.WrapH(promMetrics.Handler()))
		v1.POST("/tests", testHandler.StartTest)
		v1.GET("/tests", testHandler.ListTests)
		v1.DELETE("/tests", testHandler.ClearAllTests)
		v1.GET("/tests/summary", testHandler.GetPerformanceSummary)
		v1.GET("/tests/status", testHandler.GetStatus)
		v1.GET("/tests/metrics", testHandler.GetMetrics)
		v1.POST("/tests/stop", testHandler.StopTest)
		v1.GET("/tests/compare", testHandler.CompareTests)
		v1.GET("/tests/:id", testHandler.GetTestByID)
		v1.DELETE("/tests/:id", testHandler.DeleteTest)
		v1.GET("/tests/:id/export", testHandler.ExportTestReport)
	}

	// 8. Setup HTTP server with timeouts
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start server in a background goroutine
	go func() {
		log.Printf("Server listening on http://localhost%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()

	// 10. Graceful shutdown on interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %v. Initiating graceful shutdown...", sig)

	// Stop any active load engine workers
	_ = loadEngine.StopTest()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}

// corsMiddleware configures standard CORS headers for local frontend integration.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

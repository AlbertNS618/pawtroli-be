package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pawtroli-be/internal/api"
	"pawtroli-be/internal/firebase"
	"pawtroli-be/internal/logger"
	"pawtroli-be/internal/middleware"
	"pawtroli-be/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger first
	if err := logger.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.CloseLogger()

	// Initialize log rotation service
	// Keep logs for 30 days, maximum 50 files
	logRotationService := services.NewLogRotationService("logs", 50, 30*24*time.Hour)
	logRotationService.Start()
	defer logRotationService.Stop()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.LogInfo("Shutting down server...")
		logRotationService.Stop()
		logger.CloseLogger()
		os.Exit(0)
	}()

	logger.LogInfo("Starting Pawtroli Backend Server...")

	firebase.InitFirebase()
	api.InitHandlers()

	// Pass log rotation service to API handlers
	api.SetLogRotationService(logRotationService)

	r := mux.NewRouter()

	// Add logging middleware to all routes
	r.Use(middleware.LoggingMiddleware)

	// Routes
	api.UserRoutes(r)
	api.PetRoutes(r)
	api.ChatRoutes(r)
	api.AdminRoutes(r)

	logger.LogInfo("🚀 Server running on :8080")
	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		logger.LogErrorf("Server failed: %v", err)
		return
	}
}

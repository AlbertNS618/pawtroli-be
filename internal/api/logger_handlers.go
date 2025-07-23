package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"pawtroli-be/internal/logger"
	"pawtroli-be/internal/services"
)

var logRotationService *services.LogRotationService

func AdminRoutes(r *mux.Router) {
	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/logs", GetLogFiles).Methods("GET")
}

// SetLogRotationService sets the log rotation service for the handlers
func SetLogRotationService(lrs *services.LogRotationService) {
	logRotationService = lrs
}

// GET /admin/logs - Get list of available log files
func GetLogFiles(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger.LogInfo("GetLogFiles called")

	if logRotationService == nil {
		logger.LogError("Log rotation service not initialized")
		http.Error(w, "Log service not available", http.StatusInternalServerError)
		return
	}

	files, err := logRotationService.GetLogFilesList()
	if err != nil {
		logger.LogErrorf("Failed to get log files list: %v", err)
		http.Error(w, "Failed to get log files", http.StatusInternalServerError)
		return
	}

	logger.LogInfof("Retrieved %d log files", len(files))
	logger.LogHTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

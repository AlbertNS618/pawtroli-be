package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pawtroli-be/internal/logger"
)

// LogRotationService handles log file rotation and cleanup
type LogRotationService struct {
	logDir         string
	maxFiles       int
	maxAge         time.Duration
	rotationTicker *time.Ticker
	stopChan       chan bool
}

// NewLogRotationService creates a new log rotation service
func NewLogRotationService(logDir string, maxFiles int, maxAge time.Duration) *LogRotationService {
	return &LogRotationService{
		logDir:   logDir,
		maxFiles: maxFiles,
		maxAge:   maxAge,
		stopChan: make(chan bool),
	}
}

// Start begins the log rotation service
func (lrs *LogRotationService) Start() {
	logger.LogInfo("Starting log rotation service...")

	// Run initial cleanup
	lrs.cleanupOldLogs()

	// Set up ticker to run every hour
	lrs.rotationTicker = time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-lrs.rotationTicker.C:
				// Check if we need to rotate the current log file
				if err := logger.RotateLogFile(); err != nil {
					logger.LogErrorf("Failed to rotate log file: %v", err)
				}
				// Clean up old log files
				lrs.cleanupOldLogs()
			case <-lrs.stopChan:
				lrs.rotationTicker.Stop()
				return
			}
		}
	}()
}

// Stop stops the log rotation service
func (lrs *LogRotationService) Stop() {
	logger.LogInfo("Stopping log rotation service...")
	if lrs.rotationTicker != nil {
		lrs.stopChan <- true
	}
}

// cleanupOldLogs removes old log files based on age and count limits
func (lrs *LogRotationService) cleanupOldLogs() {
	files, err := lrs.getLogFiles()
	if err != nil {
		logger.LogErrorf("Failed to get log files for cleanup: %v", err)
		return
	}

	now := time.Now()
	var filesToDelete []string

	// Check age-based cleanup
	for _, file := range files {
		if now.Sub(file.ModTime()) > lrs.maxAge {
			filesToDelete = append(filesToDelete, file.Name())
		}
	}

	// Check count-based cleanup
	if len(files) > lrs.maxFiles {
		// Sort files by modification time (oldest first)
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})

		// Mark excess files for deletion
		excessCount := len(files) - lrs.maxFiles
		for i := 0; i < excessCount; i++ {
			filesToDelete = append(filesToDelete, files[i].Name())
		}
	}

	// Delete marked files
	for _, fileName := range filesToDelete {
		filePath := filepath.Join(lrs.logDir, fileName)
		if err := os.Remove(filePath); err != nil {
			logger.LogErrorf("Failed to delete old log file %s: %v", fileName, err)
		} else {
			logger.LogInfof("Deleted old log file: %s", fileName)
		}
	}
}

// getLogFiles returns a list of log files in the log directory
func (lrs *LogRotationService) getLogFiles() ([]os.FileInfo, error) {
	entries, err := os.ReadDir(lrs.logDir)
	if err != nil {
		return nil, err
	}

	var logFiles []os.FileInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "pawtroli_") && strings.HasSuffix(entry.Name(), ".log") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			logFiles = append(logFiles, info)
		}
	}

	return logFiles, nil
}

// GetLogFilesList returns a list of available log files with their sizes
func (lrs *LogRotationService) GetLogFilesList() ([]map[string]interface{}, error) {
	files, err := lrs.getLogFiles()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, file := range files {
		result = append(result, map[string]interface{}{
			"name":          file.Name(),
			"size":          file.Size(),
			"modTime":       file.ModTime(),
			"sizeFormatted": formatFileSize(file.Size()),
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(result, func(i, j int) bool {
		timeI := result[i]["modTime"].(time.Time)
		timeJ := result[j]["modTime"].(time.Time)
		return timeI.After(timeJ)
	})

	return result, nil
}

// formatFileSize formats file size in human readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	WarningLogger *log.Logger
	DebugLogger   *log.Logger
	logFile       *os.File
)

// LogLevel represents different log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

// InitLogger initializes the logging system with file and console output
func InitLogger() error {
	// Create log directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create log filename with current date
	logFileName := fmt.Sprintf("pawtroli_%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	// Open log file for writing (create if not exists, append if exists)
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writers to write to both file and console
	infoWriter := io.MultiWriter(os.Stdout, logFile)
	errorWriter := io.MultiWriter(os.Stderr, logFile)
	warningWriter := io.MultiWriter(os.Stdout, logFile)
	debugWriter := io.MultiWriter(os.Stdout, logFile)

	// Initialize loggers with different prefixes and flags
	InfoLogger = log.New(infoWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errorWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(warningWriter, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(debugWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Logger initialized successfully")
	return nil
}

// LogInfo logs info level messages
func LogInfo(v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Println(v...)
	}
}

// LogInfof logs formatted info level messages
func LogInfof(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

// LogError logs error level messages
func LogError(v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Println(v...)
	}
}

// LogErrorf logs formatted error level messages
func LogErrorf(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}

// LogWarning logs warning level messages
func LogWarning(v ...interface{}) {
	if WarningLogger != nil {
		WarningLogger.Println(v...)
	}
}

// LogWarningf logs formatted warning level messages
func LogWarningf(format string, v ...interface{}) {
	if WarningLogger != nil {
		WarningLogger.Printf(format, v...)
	}
}

// LogDebug logs debug level messages
func LogDebug(v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Println(v...)
	}
}

// LogDebugf logs formatted debug level messages
func LogDebugf(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Printf(format, v...)
	}
}

// LogHTTPRequest logs HTTP request details
func LogHTTPRequest(method, path, remoteAddr string, statusCode int, duration time.Duration) {
	LogInfof("HTTP %s %s from %s - Status: %d - Duration: %v",
		method, path, remoteAddr, statusCode, duration)
}

// LogFirestoreOperation logs Firestore operation details
func LogFirestoreOperation(operation, collection, docID string, success bool, duration time.Duration) {
	if success {
		LogInfof("Firestore %s operation on %s/%s successful - Duration: %v",
			operation, collection, docID, duration)
	} else {
		LogErrorf("Firestore %s operation on %s/%s failed - Duration: %v",
			operation, collection, docID, duration)
	}
}

// LogAuthOperation logs authentication operation details
func LogAuthOperation(operation, uid string, success bool) {
	if success {
		LogInfof("Auth %s operation successful for UID: %s", operation, uid)
	} else {
		LogWarningf("Auth %s operation failed for UID: %s", operation, uid)
	}
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// RotateLogFile rotates the log file if it's a new day
func RotateLogFile() error {
	newLogFileName := fmt.Sprintf("pawtroli_%s.log", time.Now().Format("2006-01-02"))
	currentLogFileName := filepath.Base(logFile.Name())

	if newLogFileName != currentLogFileName {
		LogInfo("Rotating log file...")
		CloseLogger()
		return InitLogger()
	}
	return nil
}

package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"cherkasyoblenergo-api/internal/config"
	"cherkasyoblenergo-api/internal/logger"
)

const defaultDBName = "cherkasyoblenergo.db"

func sanitizeDBPath(dbPath string) (string, error) {
	if dbPath == "" {
		dbPath = defaultDBName
	}

	cleanPath := filepath.Clean(dbPath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	decodedPath, err := url.PathUnescape(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to decode path: %w", err)
	}

	if strings.Contains(decodedPath, "..") {
		return "", fmt.Errorf("invalid database path: path traversal detected")
	}

	allowedDir := os.Getenv("DB_ALLOWED_DIR")
	if allowedDir != "" {
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve allowed directory: %w", err)
		}
		if !strings.HasPrefix(absPath, absAllowedDir+string(filepath.Separator)) {
			return "", fmt.Errorf("database path outside allowed directory")
		}
	}

	return absPath, nil
}

func ConnectDB(cfg config.Config) (*gorm.DB, error) {
	dbPath, err := sanitizeDBPath(cfg.DBName)
	if err != nil {
		return nil, fmt.Errorf("database path validation failed: %w", err)
	}

	// Configure every connection in the pool, not just a single connection.
	params := url.Values{
		"_pragma":      {"busy_timeout(5000)", "journal_mode(WAL)", "foreign_keys(1)"},
		"_time_format": {"sqlite"},
	}
	dsn := dbPath + "?" + params.Encode()
	dialector := sqlite.Open(dsn)
	log.Printf("Using SQLite database at: %s", dbPath)

	var gormLogLevel gormlogger.LogLevel
	switch strings.ToLower(cfg.LogLevel) {
	case "silent":
		gormLogLevel = gormlogger.Silent
	case "error":
		gormLogLevel = gormlogger.Error
	case "warn", "warning":
		gormLogLevel = gormlogger.Warn
	case "info":
		gormLogLevel = gormlogger.Warn
	case "debug":
		gormLogLevel = gormlogger.Info
	default:
		gormLogLevel = gormlogger.Info
	}

	newLogger := logger.NewGormLogger(gormLogLevel, time.Second)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Successfully connected to database")
	return db, nil
}

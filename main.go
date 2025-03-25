// main.go
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	ModeDev     = "dev"
	ModeProd    = "prod"
	ModePreview = "preview"
)

//go:embed web/dist
var WebContent embed.FS

var logger *log.Logger

func main() {
	mode := ModeDev
	env := ".env.local"

	// Initialize logger
	logFile := setupLogging()
	defer func() {
		if f, ok := logger.Writer().(*os.File); ok {
			f.Close()
		}
	}()

	// Mode configuration
	flag.Func("mode", "mode:dev,preview,prod", func(s string) error {
		switch s {
		case ModePreview:
			mode = ModePreview
			env = ".env.preview"
		case ModeProd:
			mode = ModeProd
			env = ".env.prod"
		}
		return nil
	})
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(env); err != nil {
		logger.Fatalf("Error loading %s file: %v", env, err)
	}

	// Clean temporary files
	if err := cleanupTempFiles(); err != nil {
		logger.Fatalf("Failed to clean temp files: %v", err)
	}

	// Setup server
	server := NewServer(mode, logFile)
	port := os.Getenv("APP_PORT")
	logger.Printf("Server starting in localhost:%s", port)
	fmt.Printf("Server starting in localhost:%s\n", port)

	if err := http.ListenAndServe(":"+port, server.Mux); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func setupLogging() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		log.Fatalf("Failed to create tmp directory: %v", err)
	}
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	logFile := filepath.Join(dir, "logs", "application.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logger = log.New(f, "", log.LstdFlags|log.Ltime)
	return logFile
}

func cleanupTempFiles() error {
	files, err := filepath.Glob(filepath.Join("tmp", "*"))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/afman42/go-llama-pdfchat-etc/utils"
)

type Server struct {
	Mux         *http.ServeMux
	Mode        string
	Broadcaster *utils.Broadcaster
	Templates   *template.Template
}

func NewServer(mode string, logFile string) *Server {
	mux := http.NewServeMux()
	broadcaster := utils.NewBroadcaster()

	// Parse templates
	tmpl, err := template.ParseFS(WebContent, "web/dist/index.html")
	if err != nil {
		logger.Fatalf("Failed to parse templates: %v", err)
	}

	server := &Server{
		Mux:         mux,
		Mode:        mode,
		Broadcaster: broadcaster,
		Templates:   tmpl,
	}

	// Setup routes
	server.setupRoutes(logFile)

	// Start background processes
	go broadcaster.Run()
	go broadcaster.TailFile(filepath.Join("logs", "application.log"), logger)

	return server
}

func (s *Server) setupRoutes(logFile string) {
	handler := utils.WrapHandlerWithLogging(http.HandlerFunc(s.indexHandler), logger)
	s.Mux.Handle("/", handler)
	s.Mux.HandleFunc("/ws", utils.HandleWebSocketConnection(s.Broadcaster, logFile, logger))

	if s.Mode == ModePreview || s.Mode == ModeProd {
		dist, err := fs.Sub(WebContent, "web/dist")
		if err != nil {
			logger.Fatalf("Failed to setup static file serving: %v", err)
		}
		s.Mux.Handle("/assets/", http.FileServer(http.FS(dist)))
		s.Mux.HandleFunc("/vite.svg", s.serveViteSVG)
	}
}

func (s *Server) serveViteSVG(w http.ResponseWriter, r *http.Request) {
	data, err := WebContent.ReadFile("web/dist/vite.svg")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		logger.Println("File not found: vite.svg")
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

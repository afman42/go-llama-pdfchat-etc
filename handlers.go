package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/afman42/go-llama-pdfchat-etc/utils"
	"github.com/jonathanhecl/chunker"
	"github.com/jonathanhecl/gollama"
)

type ChatRequest struct {
	Txt          string `json:"txt"`
	FileLocation string `json:"fileLocation"`
	ModelChat    string `json:"modelChat"`
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		logger.Println("URL Path: Not Found")
		return
	}

	setupCORS(w)

	switch r.Method {
	case http.MethodGet:
		s.handleGet(w)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func setupCORS(w http.ResponseWriter) {
	ipCors := os.Getenv("CORS_DOMAIN")
	w.Header().Set("Access-Control-Allow-Origin", ipCors)
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handleGet(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	if err := s.Templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Printf("Template execution failed: %v", err)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	const maxFileSize = 1 << 20 // 1MB
	ctx := context.Background()

	if r.URL.Query().Has("listModel") {
		s.handleListModels(w, r, ctx)
		return
	}

	if r.URL.Query().Has("upload") {
		s.handleFileUpload(w, r, maxFileSize)
		return
	}

	s.handleChatRequest(w, r, ctx)
}

func (s *Server) handleListModels(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	q := r.URL.Query().Get("listModel")
	if q != "all" {
		utils.JsonResponse(w, http.StatusInternalServerError, "Invalid listModel query")
		logger.Println("Invalid listModel query")
		return
	}

	e := gollama.New("")
	models, err := e.ListModels(ctx)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Failed to list models")
		logger.Printf("Failed to list models: %v", err)
		return
	}

	response, _ := json.Marshal(models)
	utils.JsonResponse(w, http.StatusOK, string(response))
}

func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request, maxFileSize int64) {
	q := r.URL.Query().Get("upload")
	if q != "file" {
		utils.JsonResponse(w, http.StatusInternalServerError, "Invalid upload query")
		logger.Println("Invalid upload query")
		return
	}

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Failed to parse upload")
		logger.Printf("Upload parsing error: %v", err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
	fileLocation, filename, err := utils.UploadFile(w, r, "fileLocation", logger)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
		logger.Printf("Upload error: %v", err)
		return
	}

	if strings.Contains(fileLocation, ".pdf") {
		filePath, err := utils.RunShellCMDPdf2Txt(logger, fileLocation, filename)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Printf("PDF conversion error: %v", err)
			return
		}
		fileLocation = filePath
	}

	utils.JsonResponse(w, http.StatusOK, fileLocation)
}

func (s *Server) handleChatRequest(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Invalid JSON request")
		logger.Printf("JSON decode error: %v", err)
		return
	}

	// Validate inputs
	if req.FileLocation = strings.TrimSpace(req.FileLocation); req.FileLocation == "" {
		utils.JsonResponse(w, http.StatusBadRequest, "File location required")
		return
	}
	if req.Txt = strings.TrimSpace(req.Txt); req.Txt == "" {
		utils.JsonResponse(w, http.StatusBadRequest, "Question required")
		return
	}
	if req.ModelChat = strings.TrimSpace(req.ModelChat); req.ModelChat == "" {
		utils.JsonResponse(w, http.StatusBadRequest, "Model selection required")
		return
	}

	// Setup models
	embeddingConfig := gollama.Gollama{
		ServerAddr: os.Getenv("OLLAMA_HOST"),
		ModelName:  "nomic-embed-text:latest",
		Verbose:    true,
	}
	chatConfig := gollama.Gollama{
		ServerAddr: os.Getenv("OLLAMA_HOST"),
		ModelName:  req.ModelChat,
		Verbose:    true,
	}

	embedder := gollama.NewWithConfig(embeddingConfig)
	chatter := gollama.NewWithConfig(chatConfig)

	// Verify models
	if _, err := embedder.HasModel(ctx, embeddingConfig.ModelName); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Embedding model unavailable")
		logger.Printf("Embedding model error: %v", err)
		return
	}
	if _, err := chatter.HasModel(ctx, chatConfig.ModelName); err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Chat model unavailable")
		logger.Printf("Chat model error: %v", err)
		return
	}

	// Process document
	content, err := os.ReadFile(req.FileLocation)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Failed to read file")
		logger.Printf("File read error: %v", err)
		return
	}

	// Chunk and embed
	chunks := chunker.ChunkSentences(string(content))
	embeddings, err := s.generateEmbeddings(ctx, embedder, chunks)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Embedding generation failed")
		logger.Printf("Embedding error: %v", err)
		return
	}

	// Find relevant context
	questionEmbed, err := embedder.Embedding(ctx, req.Txt)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Question embedding failed")
		logger.Printf("Question embedding error: %v", err)
		return
	}

	contexts := s.findRelevantContexts(embeddings, questionEmbed)
	if len(contexts) == 0 {
		utils.JsonResponse(w, http.StatusOK, "No relevant context found")
		return
	}

	// Generate response
	prompt := s.buildPrompt(req.Txt, contexts)
	answer, err := chatter.Chat(ctx, prompt)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, "Chat generation failed")
		logger.Printf("Chat error: %v", err)
		return
	}

	logger.Printf("Answer generated: %s", answer.Content)
	utils.JsonResponse(w, http.StatusOK, answer.Content)
}

func (s *Server) generateEmbeddings(ctx context.Context, embedder *gollama.Gollama, chunks []string) ([]struct {
	Chunk string
	Embed []float64
}, error) {
	embeddings := make([]struct {
		Chunk string
		Embed []float64
	}, 0, len(chunks))

	for _, chunk := range chunks {
		embed, err := embedder.Embedding(ctx, chunk)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, struct {
			Chunk string
			Embed []float64
		}{Chunk: chunk, Embed: embed})
	}
	return embeddings, nil
}

func (s *Server) findRelevantContexts(embeddings []struct {
	Chunk string
	Embed []float64
}, questionEmbed []float64) []string {
	const similarityThreshold = 0.65
	contexts := make([]string, 0)

	for _, embedding := range embeddings {
		similarity := gollama.CosenoSimilarity(questionEmbed, embedding.Embed)
		if similarity > similarityThreshold {
			logger.Printf("Context found: %s (Similarity: %.2f)", embedding.Chunk, similarity)
			contexts = append(contexts, embedding.Chunk)
		}
	}
	return contexts
}

func (s *Server) buildPrompt(question string, contexts []string) string {
	return fmt.Sprintf(
		"Respond to the following question using the provided context, don't add anything else:\n\n"+
			"Context:\n%s\n\nQuestion:\n%s",
		strings.Join(contexts, "\n"),
		question,
	)
}

package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/afman42/go-llama-pdfchat-etc/utils"
	"github.com/joho/godotenv"
	"github.com/jonathanhecl/chunker"
	"github.com/jonathanhecl/gollama"
)

//go:embed web/dist
var WebContent embed.FS

var (
	IpCors string
	logger *log.Logger
)

const (
	ModeDev     = "dev"
	ModeProd    = "prod"
	ModePreview = "preview"
)

func main() {
	Mode := ModeDev
	env := ".env.local"
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Ltime)
	if _, err := os.Stat("./tmp"); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir("tmp", os.ModePerm); err != nil {
				logger.Fatal(err)
			}
			if err := os.Mkdir("logs", os.ModePerm); err != nil {
				logger.Fatal(err)
			}
		}
	}

	open, err := os.OpenFile("./logs/application.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatalln(err)
	}
	logger.SetOutput(open)
	flag.Func("mode", "mode:dev,preview,prod", func(s string) error {
		if s == ModePreview {
			Mode = ModePreview
			env = ".env.preview"
		}
		if s == ModeProd {
			Mode = ModeProd
			env = ".env.prod"
		}
		return nil
	})
	flag.Parse()

	err = godotenv.Load(env)
	if err != nil {
		logger.Fatalln("Error loading " + env + " file")
	}
	IpCors = os.Getenv("CORS_DOMAIN")
	Port := os.Getenv("APP_PORT")

	files, err := filepath.Glob(utils.PathFileTemp("*"))
	if err != nil {
		logger.Fatalln(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			logger.Fatalln(err)
		}
	}
	dist, err := fs.Sub(WebContent, "web/dist")
	if err != nil {
		logger.Fatalln(err)
		return
	}
	mux := http.NewServeMux()
	handler := utils.WrapHandlerWithLogging(http.HandlerFunc(index), logger)
	mux.Handle("/", handler)
	if Mode == ModePreview || Mode == ModeProd {
		mux.Handle("/assets/", http.FileServer(http.FS(dist)))
		// Static Folder web/public
		mux.Handle("/vite.svg", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := WebContent.ReadFile("web/dist/vite.svg")
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				logger.Println("File not found: vite.svg")
				return
			}
			w.Header().Set("Content-Type", "image/svg+xml")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}))
	}
	logger.Println("Server starting in localhost:" + Port)
	fmt.Println("Server starting in localhost:" + Port)
	err = http.ListenAndServe(":"+Port, mux)
	if err != nil {
		logger.Fatalf("Something went wrong %s", err.Error())
		fmt.Println("Something went wrong", err.Error())
		os.Exit(1)
	}
}
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		logger.Println("URL Path: Not Found")
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", IpCors)
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodGet:
		var tmp, err = template.ParseFS(WebContent, "web/dist/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Println(err)
			return
		}
		w.Header().Add("Content-Type", "text/html")
		err = tmp.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Println(err)
			return
		}

	case http.MethodPost:
		const MB = 1 << 20
		ctx := context.Background()
		if listModel := r.URL.Query().Has("listModel"); listModel {
			q := r.URL.Query().Get("listModel")
			if q == "all" {
				e := gollama.New("")
				all, err := e.ListModels(ctx)
				if err != nil {
					utils.JsonResponse(w, http.StatusInternalServerError, "Something went wrong, add list models")
					logger.Println("Something went wrong, add list models")
					return
				}
				a, _ := json.Marshal(all)
				utils.JsonResponse(w, http.StatusOK, string(a))
				return
			}
			utils.JsonResponse(w, http.StatusInternalServerError, "Something went wrong, query listModel not found")
			logger.Println("Something went wrong, query listModel not found")
			return
		}
		err := r.ParseMultipartForm(1 * MB)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, "Something went wrong, upload file")
			logger.Println(err)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1*MB)

		fileLocation, err := utils.UploadFile(w, r, "file")
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Println(err)
			return
		}
		txt := strings.TrimSpace(r.FormValue("txt"))
		if txt == "" {
			utils.JsonResponse(w, http.StatusBadRequest, "Please fill input question")
			logger.Println("Please fill input question")
			return
		}
		modelEmbed := strings.TrimSpace(r.FormValue("modelEmbed"))
		if modelEmbed == "" {
			utils.JsonResponse(w, http.StatusBadRequest, "Please fill input model Embedding")
			logger.Println("Please fill input model Embedding")
			return
		}
		modelChat := strings.TrimSpace(r.FormValue("modelChat"))
		if modelChat == "" {
			utils.JsonResponse(w, http.StatusBadRequest, "please fill input model chat")
			logger.Println("please fill input model chat")
			return
		}
		fmt.Println("Embedding model:", modelEmbed)
		logger.Println("Embedding model:", modelEmbed)
		fmt.Println("Chat model:", modelChat)
		logger.Println("Chat model:", modelChat)
		configEmbedding := gollama.Gollama{
			ServerAddr: os.Getenv("OLLAMA_HOST"),
			ModelName:  modelEmbed,
			Verbose:    true,
		}
		configChat := gollama.Gollama{
			ServerAddr: os.Getenv("OLLAMA_HOST"),
			ModelName:  modelChat,
			Verbose:    true,
		}
		e := gollama.NewWithConfig(configEmbedding)
		logger.Println(e)
		_, err = e.HasModel(ctx, modelEmbed)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Println(err)
			return
		}

		c := gollama.NewWithConfig(configChat)
		logger.Println(e)
		_, err = c.HasModel(ctx, modelChat)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Println(err)
			return
		}

		f, err := os.ReadFile(fileLocation)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Println(err)
			return
		}
		text := string(f)
		fmt.Println("File ", fileLocation, "has", len(text), "bytes...")
		logger.Println("File ", fileLocation, "has", len(text), "bytes...")
		// Chunk the text
		chunks := chunker.ChunkSentences(text)
		fmt.Println("Total chunks:", len(chunks))
		logger.Println("Total chunks:", len(chunks))
		// Embed the chunks
		embeds := make([][]float64, 0)
		for _, chunk := range chunks {
			embedding, err := e.Embedding(ctx, chunk)
			if err != nil {
				utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			embeds = append(embeds, embedding)
		}
		// Save into a struct
		type tEmbedding struct {
			Chunk string
			Embed []float64
		}
		embeddings := make([]tEmbedding, 0)
		for i, embedding := range embeds {
			embeddings = append(embeddings, tEmbedding{Chunk: chunks[i], Embed: embedding})
		}
		fmt.Println("Total embeddings:", len(embeddings))
		logger.Println("Total embeddings:", len(embeddings))
		// Get the question embedding
		question_emb, _ := e.Embedding(ctx, txt)

		// Search contexts
		contexts := make([]string, 0)
		for _, embedding := range embeddings {
			similarity := gollama.CosenoSimilarity(question_emb, embedding.Embed)
			if similarity > 0.65 {
				fmt.Println("> Context:", embedding.Chunk+" (Similarity: "+fmt.Sprintf("%.2f", similarity)+")")
				logger.Println("> Context:", embedding.Chunk+" (Similarity: "+fmt.Sprintf("%.2f", similarity)+")")
				contexts = append(contexts, embedding.Chunk)
			}
		}

		if len(contexts) == 0 {
			fmt.Println("> No context found")
			logger.Println("> No context found")
			utils.JsonResponse(w, http.StatusOK, "No context found, so get another question")
			return
		}
		// Create the prompt
		prompt := "Respond to the following question using the provided context, don't add anything else:\n\n" +
			"Context:\n" + strings.Join(contexts, "\n") + "\n\nQuestion:\n" + txt

		fmt.Println("Prompt:", prompt)
		logger.Println(prompt)
		// Get the answer
		answer, err := c.Chat(ctx, prompt)
		if err != nil {
			utils.JsonResponse(w, http.StatusInternalServerError, err.Error())
			logger.Println(err)
			return
		}
		logger.Println("Answer Question: ", answer.Content)
		utils.JsonResponse(w, http.StatusOK, answer.Content)
		return
	case http.MethodOptions:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		logger.Println("You caught in method OPTIONS")
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, POST, OPTIONS")
		logger.Println("You caught in method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

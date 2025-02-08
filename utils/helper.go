package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

var fileContentTypes = map[string]string{
	"pdf": "application/pdf",
	"txt": "text/plain",
}

func StringWithCharset(length int) string {
	charset := "abcdefgABCDEFG123456"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func UploadFile(w http.ResponseWriter, r *http.Request, name string, logger *log.Logger) (string, error) {
	uploadedFile, handler, err := r.FormFile(name)
	if err != nil {
		logger.Printf("failed to get file: %s", err.Error())
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	defer uploadedFile.Close()

	_, err = getFileContentType(handler.Filename, logger)
	if err != nil {
		logger.Printf("failed to get content type for file %s: %s", handler.Filename, err.Error())
		return "", fmt.Errorf(`failed to get content type for file "%s": %w`, handler.Filename, err)
	}
	dir, err := os.Getwd()
	if err != nil {
		logger.Println("failed get directory: %w", err)
		return "", fmt.Errorf("failed get directory: %w", err)
	}

	filename := handler.Filename
	filename = fmt.Sprintf("%s%s", StringWithCharset(5), filepath.Ext(handler.Filename))
	fileLocation := filepath.Join(dir, "tmp", filename)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Printf("failed open file: %s, path: %s", err.Error(), fileLocation)
		return "", fmt.Errorf("failed open file: %w, path: %s", err, fileLocation)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, uploadedFile); err != nil {
		logger.Printf("failed system copy file: %s", err.Error())
		return "", fmt.Errorf("failed system copy file: %w", err)
	}
	return fileLocation, nil
}

func getFileContentType(fname string, logger *log.Logger) (string, error) {
	ext := filepath.Ext(fname)
	if ext == "" {
		logger.Println("file name has no extension: ", fname)
		return "", fmt.Errorf("file name has no extension: %s", fname)
	}

	ext = strings.ToLower(ext[1:])
	ct, found := fileContentTypes[ext]
	if !found {
		logger.Println("unknown file name extension: ", ext)
		return "", fmt.Errorf("unknown file name extension: %s", ext)
	}

	return ct, nil
}

func JsonResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}{
		StatusCode: statusCode,
		Message:    message,
	})
}

func PathFileTemp(filename string, logger *log.Logger) string {
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatal(err.Error())
	}
	dir = dir + "/tmp/"
	return filepath.Join(dir + filename)
}

// https://stackoverflow.com/questions/27234861/correct-way-of-getting-clients-ip-addresses-from-http-request
func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

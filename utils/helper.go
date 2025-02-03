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

func UploadFile(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	uploadedFile, handler, err := r.FormFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	defer uploadedFile.Close()

	_, err = getFileContentType(handler.Filename)
	if err != nil {
		return "", fmt.Errorf(`failed to get content type for file "%s": %w`, handler.Filename, err)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed get directory: %w", err)
	}

	filename := handler.Filename
	filename = fmt.Sprintf("%s%s", StringWithCharset(5), filepath.Ext(handler.Filename))
	fileLocation := filepath.Join(dir, "tmp", filename)
	targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("failed open file: %w, path: %s", err, fileLocation)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, uploadedFile); err != nil {
		return "", fmt.Errorf("failed system copy file: %w", err)
	}
	return fileLocation, nil
}

func getFileContentType(fname string) (string, error) {
	ext := filepath.Ext(fname)
	if ext == "" {
		return "", fmt.Errorf("file name has no extension: %s", fname)
	}

	ext = strings.ToLower(ext[1:])
	ct, found := fileContentTypes[ext]
	if !found {
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

func PathFileTemp(filename string) string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dir = dir + "/tmp/"
	return filepath.Join(dir + filename)
}

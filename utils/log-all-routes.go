package utils

import (
	"log"
	"net/http"
)

// https://ndersson.me/post/capturing_status_code_in_net_http/
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func WrapHandlerWithLogging(wrappedHandler http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Printf("--> %s %s ip user: %s", req.Method, req.URL.Path, ReadUserIP(req))

		lrw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, req)

		statusCode := lrw.statusCode
		logger.Printf("<-- %d %s ip user: %s", statusCode, http.StatusText(statusCode), ReadUserIP(req))
	})
}

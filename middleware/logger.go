package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/handler/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log, err := logger.GetLogger()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		log.Info(fmt.Sprintf("---> %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		msg := fmt.Sprintf("<--- %s %s | %d | %v", r.Method, r.URL.Path, recorder.statusCode, duration)
		if recorder.statusCode >= 500 {
			log.Error(msg)
		} else {
			log.Info(msg)
		}
	})
}

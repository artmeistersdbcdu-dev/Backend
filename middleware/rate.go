package middleware

import (
	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	token    float64
	lastSeen time.Time
}

var (
	visitor = make(map[string]*bucket)
	mu      sync.Mutex
	maxRate = 30

	rate = 10
)

func init() {
	go cleanUp()
}
func getIPAddr(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
func cleanUp() {
	for range time.Tick(5 * time.Minute) {
		mu.Lock()
		for ip, b := range visitor {
			if time.Since(b.lastSeen) > 5*time.Minute {
				delete(visitor, ip)
			}
		}
		mu.Unlock()
	}
}
func getBucket(ip string) *bucket {
	b, ok := visitor[ip]
	if !ok {
		b = &bucket{
			token:    float64(maxRate),
			lastSeen: time.Now(),
		}
		visitor[ip] = b
		return b
	}
	elapsed := time.Since(b.lastSeen).Seconds()
	b.token = math.Min(float64(maxRate), b.token+(elapsed*float64(rate)))
	b.lastSeen = time.Now()
	return b
}
func MiddlewareRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddr(r)
		mu.Lock()
		defer mu.Unlock()
		b := getBucket(ip)
		if b.token < 1 {
			handler.RespondWithError(w, http.StatusTooManyRequests, "Too many Request")
			return
		}
		b.token--
		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors   = make(map[string]*visitor)
	mu         sync.RWMutex
	rateLimit  rate.Limit = 5
	burstLimit            = 10
	rateWindow            = time.Minute // default 1 minute
)

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > rateWindow {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func getVisitor(ip string) *rate.Limiter {
	mu.RLock()
	v, exists := visitors[ip]
	mu.RUnlock()
	if !exists {
		limiter := rate.NewLimiter(rateLimit, burstLimit)
		mu.Lock()
		visitors[ip] = &visitor{limiter, time.Now()}
		mu.Unlock()
		return limiter
	}

	mu.Lock()
	v.lastSeen = time.Now()
	mu.Unlock()
	return v.limiter
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Unable to determine IP", http.StatusInternalServerError)
			return
		}

		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			clientIP = strings.Split(ip, ",")[0]
		}

		limiter := getVisitor(clientIP)
		if !limiter.Allow() {
			log.WithCtx(r.Context()).Warn().Msgf("Rate limit exceeded for IP %s", clientIP)
			response.Error(w, http.StatusTooManyRequests, "Too Many Requests", "Please slow down.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func init() {
	go cleanupVisitors()
}

func SetRateLimits(r, b int, rw time.Duration) {
	rateLimit = rate.Limit(r)
	burstLimit = b
	rateWindow = rw
}

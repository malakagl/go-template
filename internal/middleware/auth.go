package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/malakagl/kart-challenge/pkg/cache"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/db"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
	"github.com/malakagl/kart-challenge/pkg/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" { // skip auth
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("x-api-key")
		if apiKey == "" {
			log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
			return
		}

		apiKeyCached, found := apiKeyCache.Get(apiKey)
		if found {
			for _, ep := range apiKeyCached.Endpoints {
				if ep.HTTPEndpoint == r.RequestURI && ep.HTTPMethod == r.Method {
					next.ServeHTTP(w, r)
					return
				}
			}

			log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
			return
		}

		parts := strings.SplitN(apiKey, ".", 2)
		if len(parts) != 2 {
			log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
			return
		}

		clientID := parts[0]
		apiKeyDetails, err := apiKeyRepo.GetEndPoints(clientID)
		if err != nil {
			log.WithCtx(r.Context()).Error().Err(err).Msgf("internal server error %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusInternalServerError, "InternalError", http.StatusText(http.StatusInternalServerError))
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(apiKeyDetails.APIKey), []byte(parts[1])); err != nil {
			log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
			return
		}

		for _, ep := range apiKeyDetails.Endpoints {
			if ep.HTTPEndpoint == r.RequestURI && ep.HTTPMethod == r.Method {
				apiKeyCache.Put(apiKey, apiKeyDetails)
				next.ServeHTTP(w, r)
				return
			}
		}

		log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
		response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
	})
}

var (
	apiKeyRepo  *repositories.ApiKeyRepository
	apiKeyCache *cache.LRUCache[*db.APIKey]
)

func InitAuth(database *gorm.DB, cacheSize int, cacheTTL time.Duration) {
	apiKeyRepo = repositories.NewApiKeyRepository(database)
	apiKeyCache = cache.NewLRUCache[*db.APIKey](cacheSize, cacheTTL)
	go cleanupExpiredAPIKeys()
}

func cleanupExpiredAPIKeys() {
	for {
		time.Sleep(time.Minute)
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > rateWindow {
				delete(visitors, ip)
			}
		}
	}
}

package middleware

import (
	"net/http"

	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/dto/response"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" { // skip auth
			next.ServeHTTP(w, r)
			return
		}

		apiKey := r.Header.Get("x-api-key")
		if apiKey == "" || !isValidApiKey(apiKey, r.Method, r.RequestURI) {
			log.WithCtx(r.Context()).Debug().Msgf("invalid api key %s for %s %s", apiKey, r.Method, r.RequestURI)
			response.Error(w, http.StatusUnauthorized, "AuthError", http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

type accessKey struct {
	Method string
	Uri    string
}

var apiKeyMap = map[accessKey]string{
	{Method: http.MethodPost, Uri: "/orders"}: "create_order",
}

// Replace with actual API key validation logic
func isValidApiKey(key, reqMethod, reqURI string) bool {
	if v, ok := apiKeyMap[accessKey{Method: reqMethod, Uri: reqURI}]; ok {
		return v == key
	}

	return key == "apitest" // default key
}

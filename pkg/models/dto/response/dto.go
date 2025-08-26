package response

import (
	"encoding/json"
	"net/http"

	"github.com/malakagl/kart-challenge/pkg/log"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
}

func JSON(w http.ResponseWriter, code int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Msgf("error while encoding response: %v", err)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Code:    http.StatusInternalServerError,
			Type:    "error",
			Message: "error encoding response",
		})
	}
}

func Error(w http.ResponseWriter, code int, errType, message string) {
	JSON(w, code, APIResponse{
		Code:    code,
		Type:    errType,
		Message: message,
	})
}

func Success(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, APIResponse{
		Code:    status,
		Type:    "Success",
		Message: "OK",
		Data:    data,
	})
}

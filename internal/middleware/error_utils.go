package middleware

import (
	"encoding/json"
	"net/http"
)

func sendErrorResponse(w http.ResponseWriter, status int, code string, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	responseBody := struct {
		ErrorMessage *string `json:"errorMessage,omitempty"`
		ErrorCode    *string `json:"errorCode,omitempty"`
		StatusCode   *int    `json:"statusCode,omitempty"`
	}{
		ErrorMessage: StrPnc(msg),
		ErrorCode:    StrPnc(code),
		StatusCode:   IntPnc(status),
	}

	json.NewEncoder(w).Encode(responseBody)
}

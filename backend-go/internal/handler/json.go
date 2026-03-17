package handler

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error writes a JSON error response matching FastAPI's {"detail": "..."} format.
func Error(w http.ResponseWriter, status int, detail string) {
	JSON(w, status, map[string]string{"detail": detail})
}

// ValidationError writes a 422 response matching FastAPI's validation error format.
func ValidationError(w http.ResponseWriter, detail string) {
	Error(w, http.StatusUnprocessableEntity, detail)
}

// Decode reads and unmarshals a JSON request body into dst.
// Returns false and writes a 422 error if it fails.
func Decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		ValidationError(w, "Invalid request body")
		return false
	}
	return true
}

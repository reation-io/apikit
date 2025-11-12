package apikit

import (
	"encoding/json"
	"net/http"
)

// statusCoder interface for errors that include their own status code
type statusCoder interface {
	StatusCode() int
}

// WriteJSON writes a JSON response with default 200 OK status
func WriteJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteJSONWithStatus writes a JSON response with a specific status code
func WriteJSONWithStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Status already written, can't change it
		return
	}
}

// WriteError writes an error response with the given status code
func WriteError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Check if it's the custom Error type
	if apiErr, ok := err.(*Error); ok {
		json.NewEncoder(w).Encode(apiErr)
		return
	}

	// Default error format
	json.NewEncoder(w).Encode(map[string]any{
		"error": err.Error(),
	})
}

// HandleError handles errors with custom status codes
func HandleError(w http.ResponseWriter, err error) {
	if sc, ok := err.(statusCoder); ok {
		WriteError(w, err, sc.StatusCode())
		return
	}

	// Default to 500 Internal Server Error
	WriteError(w, err, http.StatusInternalServerError)
}

// HandleResponse handles both the response and error from a handler
func HandleResponse(w http.ResponseWriter, response any, err error) {
	// Handle error first
	if err != nil {
		HandleError(w, err)
		return
	}

	// Handle successful response
	WriteJSON(w, response)
}

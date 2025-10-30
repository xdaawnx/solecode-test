package http

import (
	"encoding/json"
	"net/http"
	"solecode/pkg/validator"
	"solecode/src/entities"
)

// / ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string                      `json:"error" example:"Validation failed"`
	Details []validator.ValidationError `json:"details"`
}

// notFoundHandler handles 404 errors
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, ErrorResponse{
		Error: "Endpoint not found",
	})
}

// methodNotAllowedHandler handles 405 errors
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
		Error: "Method not allowed",
	})
}

func writeValidationError(w http.ResponseWriter, message, code string) {
	writeJSON(w, http.StatusBadRequest, ValidationErrorResponse{
		Error: "Validation failed",
		Details: []validator.ValidationError{
			{
				Field:   "request",
				Message: message,
			},
		},
	})
}

func writeValidationErrors(w http.ResponseWriter, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		writeJSON(w, http.StatusBadRequest, ValidationErrorResponse{
			Error:   "Validation failed",
			Details: validationErrors,
		})
	} else {
		writeError(w, http.StatusBadRequest, err.Error())
	}
}

// writeJSON writes JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: message,
	})
}

// toUserResponse converts User entity to UserResponse
func toUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

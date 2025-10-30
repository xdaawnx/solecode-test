package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Router handles all HTTP routing
type Router struct {
	router *mux.Router
}

// NewRouter creates a new router with all routes configured
func NewRouter(userHandler *UserHandler) *Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// User routes
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	)
	// Health check
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Serve Swagger UI
	r.PathPrefix("/swagger/").Handler(swaggerHandler)

	// Serve Swagger JSON
	r.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// NotFound handler
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	// MethodNotAllowed handler
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	return &Router{
		router: r,
	}
}

// GetHandler returns the HTTP handler for the router
func (r *Router) GetHandler() http.Handler {
	return r.router
}

// healthCheck handles health check requests
func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, response)
}

package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"solecode/pkg/validator"
	uc "solecode/src/usecase"

	"github.com/gorilla/mux"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userUseCase uc.UseCases
	validator   *validator.Validator
}

func NewUserHandler(userUseCase uc.UseCases) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		validator:   validator.New(),
	}
}

// CreateUserRequest represents the request body for creating a user
// @Description Create user request
type CreateUserRequest struct {
	Name  string `json:"name" example:"John Doe" validate:"required,min=2,max=100,name"`
	Email string `json:"email" example:"john@example.com" validate:"required,email"`
}

// UserResponse represents the user response
// @Description User response
type UserResponse struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"John Doe"`
	Email     string `json:"email" example:"john@example.com"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with name and email
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User object"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ValidationErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Validate request using validator
	if err := h.validator.ValidateStruct(&req); err != nil {
		writeValidationErrors(w, err)
		return
	}
	user, err := h.userUseCase.User.CreateUser(req.Name, req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userUseCase.User.GetUser(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateUser godoc
// @Summary Update user information
// @Description Update user name and email
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body CreateUserRequest true "User object"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ValidationErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request using validator
	if err := h.validator.ValidateStruct(&req); err != nil {
		writeValidationErrors(w, err)
		return
	}

	user, err := h.userUseCase.User.UpdateUser(id, req.Name, req.Email)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "user not found":
			status = http.StatusNotFound
		case "email already exists", "invalid email format", "name is required", "email is required":
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Soft delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.userUseCase.User.DeleteUser(id); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/yourusername/notes-api/internal/domain"
	"github.com/yourusername/notes-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	validate    *validator.Validate
	logger      *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, validate *validator.Validate, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validate,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, h.logger, err)
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, h.logger, err)
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

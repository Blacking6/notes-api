package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/yourusername/notes-api/internal/domain"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	var status int
	var message string
	var details map[string]string

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		message = "resource not found"
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		message = "resource already exists"
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		message = "invalid credentials"
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		message = "validation failed"
	default:
		// Check for validator errors
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			status = http.StatusBadRequest
			message = "validation failed"
			details = make(map[string]string)
			for _, e := range validationErrs {
				details[e.Field()] = e.Tag()
			}
		} else {
			status = http.StatusInternalServerError
			message = "internal server error"
			logger.Error("unhandled error", "error", err)
		}
	}

	writeJSON(w, status, ErrorResponse{
		Error:   message,
		Details: details,
	})
}

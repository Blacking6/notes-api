package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/yourusername/notes-api/internal/domain"
	"github.com/yourusername/notes-api/internal/middleware"
	"github.com/yourusername/notes-api/internal/service"
)

type NoteHandler struct {
	noteService *service.NoteService
	validate    *validator.Validate
	logger      *slog.Logger
}

func NewNoteHandler(noteService *service.NoteService, validate *validator.Validate, logger *slog.Logger) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
		validate:    validate,
		logger:      logger,
	}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	var req domain.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, h.logger, err)
		return
	}

	note, err := h.noteService.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	noteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid note id"})
		return
	}

	note, err := h.noteService.GetByID(r.Context(), userID, noteID)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	resp, err := h.noteService.List(r.Context(), userID, page, limit)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	noteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid note id"})
		return
	}

	var req domain.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, h.logger, err)
		return
	}

	note, err := h.noteService.Update(r.Context(), userID, noteID, req)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	noteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid note id"})
		return
	}

	if err := h.noteService.Delete(r.Context(), userID, noteID); err != nil {
		writeError(w, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

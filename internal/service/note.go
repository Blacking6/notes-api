package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourusername/notes-api/internal/domain"
)

type NoteService struct {
	noteRepo domain.NoteRepository
}

func NewNoteService(noteRepo domain.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

func (s *NoteService) Create(ctx context.Context, userID uuid.UUID, req domain.CreateNoteRequest) (*domain.Note, error) {
	note := &domain.Note{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return note, nil
}

func (s *NoteService) GetByID(ctx context.Context, userID, noteID uuid.UUID) (*domain.Note, error) {
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return nil, err
	}
	if note.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return note, nil
}

func (s *NoteService) List(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.ListNotesResponse, error) {
	notes, total, err := s.noteRepo.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}
	return &domain.ListNotesResponse{
		Notes: notes,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *NoteService) Update(ctx context.Context, userID, noteID uuid.UUID, req domain.UpdateNoteRequest) (*domain.Note, error) {
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return nil, err
	}
	if note.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if req.Title != nil {
		note.Title = *req.Title
	}
	if req.Content != nil {
		note.Content = *req.Content
	}

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteService) Delete(ctx context.Context, userID, noteID uuid.UUID) error {
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return err
	}
	if note.UserID != userID {
		return domain.ErrForbidden
	}
	return s.noteRepo.Delete(ctx, noteID)
}

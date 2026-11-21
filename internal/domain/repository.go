package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByID(ctx context.Context, id uuid.UUID) (*Note, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]Note, int, error)
	Update(ctx context.Context, note *Note) error
	Delete(ctx context.Context, id uuid.UUID) error
}

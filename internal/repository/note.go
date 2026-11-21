package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourusername/notes-api/internal/domain"
)

type NoteRepo struct {
	pool *pgxpool.Pool
}

func NewNoteRepo(pool *pgxpool.Pool) *NoteRepo {
	return &NoteRepo{pool: pool}
}

func (r *NoteRepo) Create(ctx context.Context, note *domain.Note) error {
	query := `
		INSERT INTO notes (id, user_id, title, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	now := time.Now().UTC()
	note.ID = uuid.New()
	note.CreatedAt = now
	note.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		note.ID,
		note.UserID,
		note.Title,
		note.Content,
		note.CreatedAt,
		note.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}
	return nil
}

func (r *NoteRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE id = $1
	`
	var note domain.Note
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&note.ID,
		&note.UserID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get note by id: %w", err)
	}
	return &note, nil
}

func (r *NoteRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Note, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM notes WHERE user_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notes: %w", err)
	}

	query := `
		SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]domain.Note, 0, limit)
	for rows.Next() {
		var note domain.Note
		if err := rows.Scan(
			&note.ID,
			&note.UserID,
			&note.Title,
			&note.Content,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return notes, total, nil
}

func (r *NoteRepo) Update(ctx context.Context, note *domain.Note) error {
	query := `
		UPDATE notes
		SET title = $1, content = $2, updated_at = $3
		WHERE id = $4
	`
	note.UpdatedAt = time.Now().UTC()
	cmd, err := r.pool.Exec(ctx, query, note.Title, note.Content, note.UpdatedAt, note.ID)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notes WHERE id = $1`
	cmd, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

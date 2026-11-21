package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/notes-api/internal/auth"
	"github.com/yourusername/notes-api/internal/domain"
	"github.com/yourusername/notes-api/internal/service"
)

// mockUserRepo is a simple in-memory implementation for unit tests
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return domain.ErrAlreadyExists
	}
	user.ID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	tm := auth.NewTokenManager("test-secret-key-that-is-long-enough-32", 24*time.Hour)
	svc := service.NewAuthService(repo, tm)

	ctx := context.Background()

	t.Run("successful register", func(t *testing.T) {
		resp, err := svc.Register(ctx, domain.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.NotEqual(t, uuid.Nil, resp.User.ID)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := svc.Register(ctx, domain.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})

	t.Run("successful login", func(t *testing.T) {
		resp, err := svc.Login(ctx, domain.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, domain.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("unknown email", func(t *testing.T) {
		_, err := svc.Login(ctx, domain.LoginRequest{
			Email:    "unknown@example.com",
			Password: "password123",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}

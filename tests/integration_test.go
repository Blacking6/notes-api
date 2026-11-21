//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/notes-api/internal/auth"
	"github.com/yourusername/notes-api/internal/config"
	"github.com/yourusername/notes-api/internal/handler"
	"github.com/yourusername/notes-api/internal/middleware"
	"github.com/yourusername/notes-api/internal/repository"
	"github.com/yourusername/notes-api/internal/service"
)

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	// Use environment variables or defaults for test DB
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-long-enough-for-tests-32")
	os.Setenv("DB_HOST", getEnv("DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnv("DB_PORT", "5432"))
	os.Setenv("DB_USER", getEnv("DB_USER", "notes"))
	os.Setenv("DB_PASSWORD", getEnv("DB_PASSWORD", "notes"))
	os.Setenv("DB_NAME", getEnv("DB_NAME", "notes"))
	os.Setenv("DB_SSLMODE", "disable")

	cfg, err := config.Load()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL())
	require.NoError(t, err)

	userRepo := repository.NewUserRepo(pool)
	noteRepo := repository.NewNoteRepo(pool)
	tm := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration())
	authSvc := service.NewAuthService(userRepo, tm)
	noteSvc := service.NewNoteService(noteRepo)

	validate := validator.New()
	authH := handler.NewAuthHandler(authSvc, validate, nil)
	noteH := handler.NewNoteHandler(noteSvc, validate, nil)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Get("/health", handler.Health)
	r.Post("/api/v1/auth/register", authH.Register)
	r.Post("/api/v1/auth/login", authH.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(tm))
		r.Get("/api/v1/notes", noteH.List)
		r.Post("/api/v1/notes", noteH.Create)
		r.Get("/api/v1/notes/{id}", noteH.GetByID)
		r.Put("/api/v1/notes/{id}", noteH.Update)
		r.Delete("/api/v1/notes/{id}", noteH.Delete)
	})

	ts := httptest.NewServer(r)

	cleanup := func() {
		ts.Close()
		pool.Close()
	}

	return ts, cleanup
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestHealth(t *testing.T) {
	ts, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRegisterLoginAndCreateNote(t *testing.T) {
	ts, cleanup := setupTestServer(t)
	defer cleanup()

	email := "integration-" + time.Now().Format("150405") + "@example.com"
	password := "password123"

	// Register
	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var authResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&authResp))
	require.NotEmpty(t, authResp.Token)

	// Create note
	noteBody, _ := json.Marshal(map[string]string{
		"title":   "Integration test note",
		"content": "This is a test note created during integration test",
	})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/notes", bytes.NewReader(noteBody))
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	req.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
}

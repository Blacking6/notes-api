package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/yourusername/notes-api/internal/auth"
	"github.com/yourusername/notes-api/internal/config"
	"github.com/yourusername/notes-api/internal/handler"
	"github.com/yourusername/notes-api/internal/middleware"
	"github.com/yourusername/notes-api/internal/repository"
	"github.com/yourusername/notes-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Repositories
	userRepo := repository.NewUserRepo(pool)
	noteRepo := repository.NewNoteRepo(pool)

	// Auth
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration())

	// Services
	authService := service.NewAuthService(userRepo, tokenManager)
	noteService := service.NewNoteService(noteRepo)

	// Handlers
	validate := validator.New()
	authHandler := handler.NewAuthHandler(authService, validate, logger)
	noteHandler := handler.NewNoteHandler(noteService, validate, logger)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(chimw.Timeout(60 * time.Second))

	// Public routes
	r.Get("/health", handler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(tokenManager))

			r.Route("/notes", func(r chi.Router) {
				r.Get("/", noteHandler.List)
				r.Post("/", noteHandler.Create)
				r.Get("/{id}", noteHandler.GetByID)
				r.Put("/{id}", noteHandler.Update)
				r.Delete("/{id}", noteHandler.Delete)
			})
		})
	})

	server := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("starting server", "addr", cfg.HTTPAddr(), "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped")
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	return slog.New(handler)
}

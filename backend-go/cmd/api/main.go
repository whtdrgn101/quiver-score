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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/database"
	"github.com/quiverscore/backend-go/internal/handler"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/proxy"
	"github.com/quiverscore/backend-go/internal/repository"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURL := cfg.NormalizeDatabaseURL()
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("connected to database")

	r := newRouter(cfg, pool)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}

func newRouter(cfg *config.Config, pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogging)
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.Get("/health", handler.Health)

	// Create repositories
	userRepo := &repository.UserRepo{DB: pool}
	roundRepo := &repository.RoundRepo{DB: pool}
	equipmentRepo := &repository.EquipmentRepo{DB: pool}
	setupRepo := &repository.SetupRepo{DB: pool}
	sightMarkRepo := &repository.SightMarkRepo{DB: pool}
	classificationRepo := &repository.ClassificationRepo{DB: pool}
	scoringRepo := &repository.ScoringRepo{DB: pool}

	// Create handlers
	authHandler := &handler.AuthHandler{Users: userRepo, Cfg: cfg}
	usersHandler := &handler.UsersHandler{Users: userRepo, Cfg: cfg}
	roundsHandler := &handler.RoundsHandler{Rounds: roundRepo, Cfg: cfg}
	equipmentHandler := &handler.EquipmentHandler{Equipment: equipmentRepo, Cfg: cfg}
	setupsHandler := &handler.SetupsHandler{Setups: setupRepo, Cfg: cfg}
	sightMarksHandler := &handler.SightMarksHandler{SightMarks: sightMarkRepo, Cfg: cfg}
	classificationsHandler := &handler.ClassificationsHandler{Classifications: classificationRepo, Cfg: cfg}
	scoringHandler := &handler.ScoringHandler{Scoring: scoringRepo, Cfg: cfg}
	sharingHandler := &handler.SharingHandler{Scoring: scoringRepo, Users: userRepo, Rounds: roundRepo, Cfg: cfg}

	r.Route("/api/v1/auth", authHandler.Routes)
	r.Route("/api/v1/rounds", roundsHandler.Routes)
	r.Route("/api/v1/equipment", equipmentHandler.Routes)
	r.Route("/api/v1/setups", setupsHandler.Routes)
	r.Route("/api/v1/sight-marks", sightMarksHandler.Routes)
	r.Route("/api/v1/sessions", scoringHandler.Routes)
	r.Route("/api/v1/share", sharingHandler.Routes)

	// Mount users/me as a group so we can add sub-routes
	r.Route("/api/v1/users/me", func(ur chi.Router) {
		ur.With(middleware.RequireAuth(cfg.SecretKey)).Get("/", usersHandler.GetMe)
		ur.With(middleware.RequireAuth(cfg.SecretKey)).Patch("/", usersHandler.UpdateMe)
		ur.With(middleware.RequireAuth(cfg.SecretKey)).Post("/avatar", usersHandler.UploadAvatar)
		ur.With(middleware.RequireAuth(cfg.SecretKey)).Post("/avatar-url", usersHandler.UploadAvatarFromURL)
		ur.With(middleware.RequireAuth(cfg.SecretKey)).Delete("/avatar", usersHandler.DeleteAvatar)
		ur.Route("/classifications", classificationsHandler.Routes)
	})

	// Public profile (no auth required)
	r.Get("/api/v1/users/{username}", usersHandler.GetPublicProfile)

	// Proxy everything else to the Python API
	pythonProxy := proxy.New(cfg.PythonAPIURL)
	r.NotFound(pythonProxy.ServeHTTP)

	slog.Info("proxying unhandled routes to Python API", "url", cfg.PythonAPIURL)

	return r
}

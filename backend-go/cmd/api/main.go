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

	authHandler := &handler.AuthHandler{DB: pool, Cfg: cfg}
	usersHandler := &handler.UsersHandler{DB: pool, Cfg: cfg}
	roundsHandler := &handler.RoundsHandler{DB: pool, Cfg: cfg}
	equipmentHandler := &handler.EquipmentHandler{DB: pool, Cfg: cfg}
	setupsHandler := &handler.SetupsHandler{DB: pool, Cfg: cfg}

	r.Route("/api/v1/auth", authHandler.Routes)
	r.Route("/api/v1/rounds", roundsHandler.Routes)
	r.Route("/api/v1/equipment", equipmentHandler.Routes)
	r.Route("/api/v1/setups", setupsHandler.Routes)

	// Mount users/me directly (not as a subrouter) so that deeper paths
	// like /api/v1/users/me/classifications/current fall through to the proxy.
	r.With(middleware.RequireAuth(cfg.SecretKey)).Get("/api/v1/users/me", usersHandler.GetMe)

	// Proxy everything else to the Python API
	pythonProxy := proxy.New(cfg.PythonAPIURL)
	r.NotFound(pythonProxy.ServeHTTP)

	slog.Info("proxying unhandled routes to Python API", "url", cfg.PythonAPIURL)

	return r
}

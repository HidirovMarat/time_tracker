package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"log/slog"
	tCreate "time_tracker/internal/http-server/handlers/task/create"
	tGetUT "time_tracker/internal/http-server/handlers/task/getUserTasks"
	tStart "time_tracker/internal/http-server/handlers/task/start"
	tStop "time_tracker/internal/http-server/handlers/task/stop"
	uCreate "time_tracker/internal/http-server/handlers/user/create"

	uDelete "time_tracker/internal/http-server/handlers/user/delete"
	uGet "time_tracker/internal/http-server/handlers/user/get"
	uUpdate "time_tracker/internal/http-server/handlers/user/update"

	"time_tracker/internal/request/info"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"time_tracker/internal/config"
	mwLogger "time_tracker/internal/http-server/middleware/logger"
	"time_tracker/internal/lib/logger/handlers/slogpretty"
	"time_tracker/internal/lib/logger/sl"
	"time_tracker/internal/storage/post"

	_ "time_tracker/docs" // Импортируйте ваши Swagger доки

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/swaggo/swag"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// @title Swagger Example API
// @version 7.0
// @description This is a sample server.
// @host localhost:8082
// @BasePath /

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting time_tracker",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
	if err := runMigrations(cfg.StoragePath); err != nil {
		log.Error("failed to run migrations", sl.Err(err))
		os.Exit(1)
	}
	log.Info("start migrations")

	storage, err := post.NewPG(context.Background(), cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	infoS := info.NewRI()

	router := chi.NewRouter()
	/*
		base/
	*/
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/user", uGet.New(context.Background(), log, storage))
	router.Delete("/user", uDelete.New(context.Background(), log, storage))
	router.Post("/user", uCreate.New(context.Background(), log, storage, infoS, cfg.Address))
	router.Patch("/user", uUpdate.New(context.Background(), log, storage))

	router.Post("/task", tCreate.New(context.Background(), log, storage))
	router.Get("/task/task-time", tGetUT.New(context.Background(), log, storage))
	router.Put("/task/start", tStart.New(context.Background(), log, storage))
	router.Put("/task/stop", tStop.New(context.Background(), log, storage))

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://db/migrations",
		databaseURL,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

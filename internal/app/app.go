package app

import (
	"context"
	"fmt"
	"time"

	httpDelivery "github.com/exPriceD/pr-reviewer-service/internal/delivery/http"
	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/handler"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/transaction"
	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/config"
	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/database"
	prRepo "github.com/exPriceD/pr-reviewer-service/internal/infrastructure/database/pull_request"
	teamRepo "github.com/exPriceD/pr-reviewer-service/internal/infrastructure/database/team"
	userRepo "github.com/exPriceD/pr-reviewer-service/internal/infrastructure/database/user"
	infraLogger "github.com/exPriceD/pr-reviewer-service/internal/infrastructure/logger"
	"github.com/exPriceD/pr-reviewer-service/internal/usecase"
)

// App содержит все зависимости приложения
type App struct {
	Config    *config.Config
	Logger    logger.Logger
	DB        *database.PostgresDB
	TxManager transaction.Manager

	// Repositories
	UserRepository        *userRepo.Repository
	TeamRepository        *teamRepo.Repository
	PullRequestRepository *prRepo.Repository

	// Use Cases
	UserUseCase        *usecase.UserUseCase
	TeamUseCase        *usecase.TeamUseCase
	PullRequestUseCase *usecase.PullRequestUseCase
	StatisticsUseCase  *usecase.StatisticsUseCase

	// HTTP Server
	HTTPServer *httpDelivery.Server
}

// Build создает и инициализирует все компоненты приложения
func Build() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log := infraLogger.NewSlogLogger(infraLogger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})

	log.Info("Starting PR Reviewer Service",
		"version", "1.0.0",
		"env", cfg.Server.Host,
		"log_level", cfg.Logger.Level,
	)

	log.Info("Connecting to database",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.DBName,
	)

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Successfully connected to database")

	txManager := database.NewTransactionManager(db)

	log.Info("Transaction Manager initialized")

	userRepository := userRepo.NewRepository(db.DB(), db.Getter())
	teamRepository := teamRepo.NewRepository(db.DB(), db.Getter())
	pullRequestRepository := prRepo.NewRepository(db.DB(), db.Getter())

	log.Info("Repositories initialized")

	reviewerSelector := usecase.NewReviewerSelector(userRepository, pullRequestRepository)

	userUseCase := usecase.NewUserUseCase(userRepository, pullRequestRepository, log)
	teamUseCase := usecase.NewTeamUseCase(txManager, teamRepository, userRepository, log)
	pullRequestUseCase := usecase.NewPullRequestUseCase(txManager, pullRequestRepository, userRepository, reviewerSelector, log)
	statisticsUseCase := usecase.NewStatisticsUseCase(pullRequestRepository, userRepository, log)

	log.Info("Use Cases initialized")

	teamHandler := handler.NewTeamHandler(teamUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	pullRequestHandler := handler.NewPullRequestHandler(pullRequestUseCase)
	statisticsHandler := handler.NewStatisticsHandler(statisticsUseCase)

	router := httpDelivery.NewRouter(teamHandler, userHandler, pullRequestHandler, statisticsHandler, log, int64(cfg.Server.MaxBodySize))
	chiRouter := router.Setup()

	httpServer := httpDelivery.NewServer(cfg.Server, chiRouter)
	log.Info("HTTP Server initialized", "address", httpServer.Address())

	return &App{
		Config:                cfg,
		Logger:                log,
		DB:                    db,
		TxManager:             txManager,
		UserRepository:        userRepository,
		TeamRepository:        teamRepository,
		PullRequestRepository: pullRequestRepository,
		UserUseCase:           userUseCase,
		TeamUseCase:           teamUseCase,
		PullRequestUseCase:    pullRequestUseCase,
		StatisticsUseCase:     statisticsUseCase,
		HTTPServer:            httpServer,
	}, nil
}

// Shutdown корректно завершает работу приложения
func (a *App) Shutdown() error {
	a.Logger.Info("Shutting down application...")

	if a.HTTPServer != nil {
		shutdownTimeout := time.Duration(a.Config.Server.ShutdownTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := a.HTTPServer.Shutdown(ctx); err != nil {
			a.Logger.Error("Error shutting down HTTP server", "error", err)
			return err
		}
		a.Logger.Info("HTTP Server stopped")
	}

	if err := a.DB.Close(); err != nil {
		a.Logger.Error("Error closing database connection", "error", err)
		return err
	}

	a.Logger.Info("Application stopped successfully")
	return nil
}

// Start запускает HTTP сервер
func (a *App) Start() error {
	if a.HTTPServer == nil {
		return fmt.Errorf("HTTP server is not initialized")
	}

	a.Logger.Info("Starting HTTP server", "address", a.HTTPServer.Address())
	return a.HTTPServer.Start()
}

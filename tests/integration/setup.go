package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/exPriceD/pr-reviewer-service/internal/app"
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

func createTestDB(ctx context.Context, cfg config.DatabaseConfig) (*database.PostgresDB, error) {
	//nolint:contextcheck
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if err := db.DB().PingContext(ctx); err == nil {
			break
		}
		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to connect to database after %d retries", maxRetries)
		}
		//nolint:forbidigo // Таймаут ожидания можно игнорировать в тестовой среде
		time.Sleep(time.Second)
	}

	return db, nil
}

func createTestTxManager(db *database.PostgresDB) transaction.Manager {
	return database.NewTransactionManager(db)
}

type testRepositories struct {
	UserRepo *userRepo.Repository
	TeamRepo *teamRepo.Repository
	PRRepo   *prRepo.Repository
}

func createTestRepositories(db *database.PostgresDB) testRepositories {
	return testRepositories{
		UserRepo: userRepo.NewRepository(db.DB(), db.Getter()),
		TeamRepo: teamRepo.NewRepository(db.DB(), db.Getter()),
		PRRepo:   prRepo.NewRepository(db.DB(), db.Getter()),
	}
}

type testUseCases struct {
	UserUseCase        *usecase.UserUseCase
	TeamUseCase        *usecase.TeamUseCase
	PullRequestUseCase *usecase.PullRequestUseCase
	StatisticsUseCase  *usecase.StatisticsUseCase
}

func createTestUseCases(txManager transaction.Manager, repos testRepositories, log logger.Logger) testUseCases {
	reviewerSelector := usecase.NewReviewerSelector(repos.UserRepo, repos.PRRepo)

	return testUseCases{
		UserUseCase:        usecase.NewUserUseCase(repos.UserRepo, repos.PRRepo, log),
		TeamUseCase:        usecase.NewTeamUseCase(txManager, repos.TeamRepo, repos.UserRepo, log),
		PullRequestUseCase: usecase.NewPullRequestUseCase(txManager, repos.PRRepo, repos.UserRepo, reviewerSelector, log),
		StatisticsUseCase:  usecase.NewStatisticsUseCase(repos.PRRepo, repos.UserRepo, log),
	}
}

type testHandlers struct {
	TeamHandler        *handler.TeamHandler
	UserHandler        *handler.UserHandler
	PullRequestHandler *handler.PullRequestHandler
	StatisticsHandler  *handler.StatisticsHandler
}

func createTestHandlers(useCases testUseCases) testHandlers {
	return testHandlers{
		TeamHandler:        handler.NewTeamHandler(useCases.TeamUseCase),
		UserHandler:        handler.NewUserHandler(useCases.UserUseCase),
		PullRequestHandler: handler.NewPullRequestHandler(useCases.PullRequestUseCase),
		StatisticsHandler:  handler.NewStatisticsHandler(useCases.StatisticsUseCase),
	}
}

func createTestRouter(handlers testHandlers, log logger.Logger, maxBodySize int64) *httpDelivery.Router {
	return httpDelivery.NewRouter(
		handlers.TeamHandler,
		handlers.UserHandler,
		handlers.PullRequestHandler,
		handlers.StatisticsHandler,
		log,
		maxBodySize,
	)
}

func createTestHTTPServer(cfg config.ServerConfig, router *httpDelivery.Router) *httpDelivery.Server {
	return httpDelivery.NewServer(cfg, router.Setup())
}

func buildTestApp(ctx context.Context, cfg *config.Config) (*app.App, error) {
	log := infraLogger.NewSlogLogger(infraLogger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})

	db, err := createTestDB(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	txManager := createTestTxManager(db)
	repos := createTestRepositories(db)
	useCases := createTestUseCases(txManager, repos, log)
	handlers := createTestHandlers(useCases)
	router := createTestRouter(handlers, log, int64(cfg.Server.MaxBodySize))
	httpServer := createTestHTTPServer(cfg.Server, router)

	return &app.App{
		Config:                cfg,
		Logger:                log,
		DB:                    db,
		TxManager:             txManager,
		UserRepository:        repos.UserRepo,
		TeamRepository:        repos.TeamRepo,
		PullRequestRepository: repos.PRRepo,
		UserUseCase:           useCases.UserUseCase,
		TeamUseCase:           useCases.TeamUseCase,
		PullRequestUseCase:    useCases.PullRequestUseCase,
		StatisticsUseCase:     useCases.StatisticsUseCase,
		HTTPServer:            httpServer,
	}, nil
}

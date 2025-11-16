package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/handler"
	"github.com/exPriceD/pr-reviewer-service/internal/delivery/http/middleware"
	"github.com/exPriceD/pr-reviewer-service/internal/domain/logger"
)

// Router настраивает HTTP роутер и middleware
type Router struct {
	teamHandler        *handler.TeamHandler
	userHandler        *handler.UserHandler
	pullRequestHandler *handler.PullRequestHandler
	statisticsHandler  *handler.StatisticsHandler
	logger             logger.Logger
	maxBodySize        int64
}

// NewRouter создает новый Router
func NewRouter(
	teamHandler *handler.TeamHandler,
	userHandler *handler.UserHandler,
	pullRequestHandler *handler.PullRequestHandler,
	statisticsHandler *handler.StatisticsHandler,
	logger logger.Logger,
	maxBodySize int64,
) *Router {
	return &Router{
		teamHandler:        teamHandler,
		userHandler:        userHandler,
		pullRequestHandler: pullRequestHandler,
		statisticsHandler:  statisticsHandler,
		logger:             logger,
		maxBodySize:        maxBodySize,
	}
}

// Setup настраивает роутер и возвращает http.Handler
func (r *Router) Setup() *chi.Mux {
	router := chi.NewRouter()

	router.Use(chimw.RequestID)
	router.Use(middleware.RequestID)
	router.Use(middleware.LimitBodySize(r.maxBodySize))
	router.Use(middleware.Logger(r.logger))
	router.Use(middleware.Recovery(r.logger))
	router.Use(chimw.RealIP)
	router.Use(chimw.NoCache)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		//nolint:gosec
		_, _ = w.Write([]byte("OK"))
	})

	r.teamHandler.RegisterRoutes(router)
	r.userHandler.RegisterRoutes(router)
	r.pullRequestHandler.RegisterRoutes(router)
	r.statisticsHandler.RegisterRoutes(router)

	return router
}

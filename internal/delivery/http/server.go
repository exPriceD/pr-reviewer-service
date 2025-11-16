package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/exPriceD/pr-reviewer-service/internal/infrastructure/config"
)

// Server представляет HTTP сервер
type Server struct {
	httpServer *http.Server
}

// NewServer создает новый HTTP сервер
func NewServer(cfg config.ServerConfig, handler http.Handler) *Server {
	if handler == nil {
		panic("HTTP handler cannot be nil")
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return &Server{
		httpServer: &http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:    time.Duration(cfg.IdleTimeout) * time.Second,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown останавливает HTTP сервер gracefully
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Address возвращает адрес сервера
func (s *Server) Address() string {
	return s.httpServer.Addr
}

// Handler возвращает HTTP handler для использования в тестах
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

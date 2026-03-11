package server

import (
	"context"
	"learning-go/internal/infrastructure/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, handler *gin.Engine) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           ":" + cfg.AppPort,
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

func (server *Server) Run() error {
	return server.httpServer.ListenAndServe()
}

func (server *Server) Addr() string {
	return server.httpServer.Addr
}

func (server *Server) Shutdown(ctx context.Context) error {
	return server.httpServer.Shutdown(ctx)
}

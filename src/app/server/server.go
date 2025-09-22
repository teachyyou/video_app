package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServerParams struct {
	fx.In
	Logger *zap.Logger
	Router *gin.Engine
}

func NewHTTPServer(p ServerParams) *http.Server {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           p.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

func RegisterLifecycle(lc fx.Lifecycle, log *zap.Logger, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting server", zap.String("address", server.Addr))
			go func() {
				_ = server.ListenAndServe()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping server", zap.String("address", server.Addr))
			return server.Shutdown(ctx)
		},
	})
}

func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

var Module = fx.Module("httpserver",
	fx.Provide(NewLogger, NewRouter, NewHTTPServer),
	fx.Invoke(RegisterLifecycle),
)

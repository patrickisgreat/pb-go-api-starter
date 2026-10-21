package appserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickisgreat/pb-go-api-starter/internal/logger"
)

const TEN_SECONDS time.Duration = 10 * time.Second

type StartParams struct {
	Addr    string
	Handler http.Handler
}

func Start(ctx context.Context, sp *StartParams) {
	log := logger.Get(ctx)

	if sctx == nil {
		err := errors.New("server context must be initialized before start")
		log.Error("Failed to start server", slog.Any("err", err))
		panic(err)
	}

	server := &http.Server{
		Addr:    sp.Addr,
		Handler: sp.Handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("Failed to start server", slog.Any("err", err))
			panic(err)
		}
	}()

	log.Info("Server started", "addr", sp.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for SIGINT or SIGTERM
	<-stop

	// Give server ten seconds to shutdown gracefully
	ctxWithTimeout, cancel := context.WithTimeout(ctx, TEN_SECONDS)
	defer cancel()

	if err := server.Shutdown(ctxWithTimeout); err != nil {
		log.Error("Failed to shut down server", slog.Any("err", err))
	}
	err := shutdownOtel(ctxWithTimeout)
	if err != nil {
		log.Error("Failed to shut down telemetry", slog.Any("err", err))
	}

	log.Info("Server stopped")
}

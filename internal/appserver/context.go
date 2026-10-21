package appserver

import (
	"context"
	"log"
	"sync"

	"github.com/patrickisgreat/pb-go-api-starter/internal/config"
	"github.com/patrickisgreat/pb-go-api-starter/internal/logger"
	"github.com/patrickisgreat/pb-go-api-starter/internal/quoterepo"
	"github.com/patrickisgreat/pb-go-api-starter/internal/quoteservice"
	"github.com/patrickisgreat/pb-go-api-starter/internal/telemetry"
)

type ServerContext struct {
	QuoteRepository *quoterepo.QuoteRepository
	QuoteService    *quoteservice.QuoteService
}

var sctx *ServerContext
var shutdownOtel func(context.Context) error
var initOnce sync.Once

func InitContext(ctx context.Context) {
	initOnce.Do(func() {
		var err error
		shutdownOtel, err = telemetry.Setup(ctx, telemetry.WithECSDetector())
		if err != nil {
			log.Fatalf("Failed to set up telemetry: %v", err)
		}
		config.Load(ctx)
		logger.Init(logger.Options{
			Level:        config.Config.LogLevel,
			Pretty:       config.Config.LogPretty,
			Service:      config.Config.AppName,
			ExportToOTel: !telemetry.Disabled(),
		})
		sctx = newContext()
	})
}

func GetServerContext() *ServerContext {
	if sctx == nil {
		log.Fatal("Server context was not initialized")
	}
	return sctx
}

func newContext() *ServerContext {
	quoteRepo := quoterepo.NewQuoteRepository()
	quoteService := quoteservice.NewQuoteService(quoteRepo)

	return &ServerContext{
		QuoteRepository: quoteRepo,
		QuoteService:    quoteService,
	}
}

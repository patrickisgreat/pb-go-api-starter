package config

import (
	"context"
	"log/slog"

	"github.com/patrickisgreat/pb-go-api-starter/internal/baseconfig"
	"github.com/patrickisgreat/pb-go-api-starter/internal/logger"
)

type config struct {
	baseconfig.Base

	LogPretty bool `envconfig:"LOG_PRETTY" default:"false"`
}

var Config config

func Load(ctx context.Context) {
	err := baseconfig.Load(&Config)
	if err != nil {
		logger.Get(ctx).Error("failed loading app config", slog.Any("err", err))
		panic("failed loading app config")
	}
}

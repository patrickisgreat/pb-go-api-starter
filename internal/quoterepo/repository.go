package quoterepo

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"math/rand"

	"github.com/patrickisgreat/pb-go-api-starter/internal/logger"
)

//go:embed quotes.json
var f embed.FS

type QuoteRepository struct {
	Quotes []string `json:"quotes"`
}

func NewQuoteRepository() *QuoteRepository {
	quotes := &QuoteRepository{}
	data, err := f.ReadFile("quotes.json")
	if err == nil {
		_ = json.Unmarshal(data, quotes)
	}
	return quotes
}

func (cr *QuoteRepository) Gimme(ctx context.Context) (string, error) {
	log := logger.Get(ctx)
	quoteCount := len(cr.Quotes)
	if quoteCount == 0 {
		err := &NoQuotesError{}
		log.Error("no quotes available", slog.Any("err", err))
		return "", err
	}
	ix := rand.Intn(len(cr.Quotes))
	return cr.Quotes[ix], nil
}

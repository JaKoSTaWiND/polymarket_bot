package actors_market

import (
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/entities"
)

// MarketActor представляет актор одного рынка.
type MarketActor struct {
	Logger    *slog.Logger // Логгер для событий рынка.
	Market    entities.Market
	Positions []entities.Position
	// Orders []entities.Order (в будущем)
}

// CreateMarketActor возвращает актор рынка с переданным логгером logger.
func CreateMarketActor(logger *slog.Logger) *MarketActor {
	return &MarketActor{Logger: logger}
}

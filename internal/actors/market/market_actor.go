package actors_market

import (
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/entities"
)

// MarketActor представляет актор одного рынка.
type MarketActor struct {
	Logger    *slog.Logger        // Логгер для событий рынка.
	Market    entities.Market     // Данные рынка и его исходов.
	Positions []entities.Position // Локальные позиции обоих исходов.
	// Orders []entities.Order (в будущем)
}

// CreateMarketActor создаёт актор для market с журналом logger и пустым слайсом позиций.
func CreateMarketActor(logger *slog.Logger, market entities.Market) *MarketActor {
	return &MarketActor{
		Logger:    logger,
		Market:    market,
		Positions: []entities.Position{},
	}
}

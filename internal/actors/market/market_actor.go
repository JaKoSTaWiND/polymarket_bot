package actors_market

import "log/slog"

// MarketActor представляет актор одного рынка.
type MarketActor struct {
	Logger *slog.Logger // Логгер для событий рынка.
}

// CreateMarketActor возвращает актор рынка с переданным логгером logger.
func CreateMarketActor(logger *slog.Logger) *MarketActor {
	return &MarketActor{Logger: logger}
}

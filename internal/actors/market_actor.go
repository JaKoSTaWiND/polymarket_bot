package actors

import "log/slog"

// MarketActor представляет актор одного рынка.
type MarketActor struct {
	Logger *slog.Logger // Логгер для событий рынка.
}

// NewMarketActor возвращает актор рынка с переданным логгером logger.
func NewMarketActor(logger *slog.Logger) *MarketActor {
	return &MarketActor{Logger: logger}
}

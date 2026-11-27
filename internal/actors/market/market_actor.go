package actors_market

import (
	"encoding/json"
	"log/slog"
)

// PositionsByToken хранит исходные позиции одного рынка по ID токена исхода.
type PositionsByToken map[string]json.RawMessage

// MarketActor представляет актор одного рынка.
type MarketActor struct {
	Logger    *slog.Logger     // Логгер для событий рынка.
	Positions PositionsByToken // Позиции кошелька на момент запуска, отдельно для каждого исхода.
}

// CreateMarketActor возвращает актор рынка с переданным логгером logger.
func CreateMarketActor(logger *slog.Logger) *MarketActor {
	return &MarketActor{Logger: logger}
}

package di_states

import (
	"errors"

	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"go.uber.org/fx"
)

// CreateGeneralState генерирует BotID, задаёт статус STARTED и создаёт карту акторов с лимитом settings.MaxTrackingMarkets.
// При nil settings лимит равен нулю. Отрицательный лимит или сбой генерации BotID дают ошибку.
func CreateGeneralState(settings *internal.BootstrapSettings) (*states.GeneralState, error) {
	var limit int
	if settings != nil {
		limit = settings.MaxTrackingMarkets
	}
	if limit < 0 {
		return nil, errors.New("MaxTrackingMarkets должен быть неотрицательным")
	}
	botID, err := states.GenerateBotID()
	if err != nil {
		return nil, err
	}
	return &states.GeneralState{
		BotID:                     botID,
		BotStatus:                 lifecycle.STARTED,
		MarketActorsByConditionID: make(map[string]actors_market.MarketActor, limit),
		MaxTrackingMarkets:        limit,
	}, nil
}

func GeneralStateProvider() fx.Option {
	return fx.Provide(CreateGeneralState)
}

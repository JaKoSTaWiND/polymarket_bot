package di_usecases

import (
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	usecases_lifecycle "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/lifecycle"
	usecases_sdk "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/sdk"
	"go.uber.org/fx"
)

func CreateLifecycle(logger *slog.Logger, state *states.GeneralState) lifecycle.Lifecycle {
	return lifecycle.Lifecycle{
		Logger:       logger,
		GeneralState: state,
	}
}

func CreateStartBotUC(
	logger *slog.Logger,
	state *states.GeneralState,
	botLifecycle lifecycle.Lifecycle,
	subscribe *usecases_sdk.SubscribeUserChannelUC,
) *usecases_lifecycle.StartBotUC {
	return &usecases_lifecycle.StartBotUC{
		Logger:                 logger,
		GeneralState:           state,
		Lifecycle:              botLifecycle,
		SubscribeUserChannelUC: subscribe,
	}
}

func LifecycleUCsProvider() fx.Option {
	return fx.Provide(CreateLifecycle, CreateStartBotUC)
}

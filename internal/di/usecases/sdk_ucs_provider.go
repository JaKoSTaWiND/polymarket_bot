package di_usecases

import (
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/sdk"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	usecases_sdk "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/sdk"
	"go.uber.org/fx"
)

func CreateSubscribeUserChannelUC(logger *slog.Logger, apis *states.ApisState, client *sdk.SDKClient) *usecases_sdk.SubscribeUserChannelUC {
	return &usecases_sdk.SubscribeUserChannelUC{
		Logger:    logger,
		ApisState: apis,
		Client:    client,
	}
}

func SDKUCsProvider() fx.Option {
	return fx.Provide(CreateSubscribeUserChannelUC)
}

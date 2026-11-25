package di_sdk

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal/sdk"
	"go.uber.org/fx"
)

func SDKClientProvider() fx.Option {
	return fx.Provide(sdk.CreateSDKClient)
}

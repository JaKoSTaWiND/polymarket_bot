package di_usecases

import (
	"log/slog"
	"net/http"

	"github.com/JaKoSTaWiND/polymarket_bot/internal"
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

// CreateGetBalanceUC создаёт сценарий чтения баланса для адреса из settings
// с журналом logger и HTTP-клиентом client. Результат записывается в общее состояние state.
func CreateGetBalanceUC(logger *slog.Logger, client *http.Client, settings *internal.BootstrapSettings, state *states.GeneralState) *usecases_sdk.GetBalanceUC {
	return &usecases_sdk.GetBalanceUC{
		Logger:            logger,
		Client:            client,
		PolyWalletAddress: settings.PolyWalletAddress,
		GeneralState:      state,
	}
}

// SDKUCsProvider регистрирует сценарии SDK в контейнере Fx.
func SDKUCsProvider() fx.Option {
	return fx.Provide(CreateSubscribeUserChannelUC, CreateGetBalanceUC)
}

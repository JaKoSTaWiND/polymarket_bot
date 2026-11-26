package di_states

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"go.uber.org/fx"
)

func CreateApisState(BootstrapSettings *internal.BootstrapSettings) (*states.ApisState, error) {
	return &states.ApisState{
		PolyRelayerApiKey: BootstrapSettings.PolyRelayerApiKey,
		PolyRelayerApiKeyAddress: BootstrapSettings.PolyRelayerApiKeyAddress,
		PolyApiKey: BootstrapSettings.PolyApiKey,
		PolySecret: BootstrapSettings.PolySecret,
		PolyPassphrase: BootstrapSettings.PolyPassphrase,
	}, nil
}

func ApisStateProvider() fx.Option {
	return fx.Provide(CreateApisState)
}
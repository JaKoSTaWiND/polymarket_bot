package di_states

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"go.uber.org/fx"
)


func CreateGeneralState(BootstrapSettings *internal.BootstrapSettings) (*states.GeneralState, error) {
	BotID, err := states.GenerateBotID()
	if err != nil {
		return nil, err
	}
	return &states.GeneralState{
		BotID:     BotID,
		BotStatus: string(lifecycle.STARTED),
	}, nil
}

func GeneralStateProvider() fx.Option {
	return fx.Provide(CreateGeneralState)
}

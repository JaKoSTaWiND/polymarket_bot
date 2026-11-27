package di_sdk

import (
	sdk_profile "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/profile"
	"go.uber.org/fx"
)

func ProfileProvider() fx.Option {
	return fx.Supply(
		sdk_profile.ListPositionsForAUser,
	)
}

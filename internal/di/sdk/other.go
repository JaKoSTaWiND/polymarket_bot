package di_sdk

import (
	sdk_other "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/other"
	"go.uber.org/fx"
)

func OtherProvider() fx.Option {
	return fx.Supply(
		sdk_other.GetBalance,
	)
}

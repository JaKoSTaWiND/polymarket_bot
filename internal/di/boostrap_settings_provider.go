package di

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"go.uber.org/fx"
)

// BootstrapSettingsProvider передаёт настройки запуска в конструкторы Fx.
func BootstrapSettingsProvider(BootstrapSettings *internal.BootstrapSettings) fx.Option {
	return fx.Supply(BootstrapSettings)
}

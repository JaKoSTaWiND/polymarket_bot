package di_observability

import (
	"log/slog"
	"os"

	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	observability_logging "github.com/JaKoSTaWiND/polymarket_bot/internal/observability/logging"
	"go.uber.org/fx"
)

const ServiceName = "Trading Bot"

func CreateLogger(Lifecycle fx.Lifecycle, Settings *internal.BootstrapSettings) (*slog.Logger, error) {
	Configuration := observability_logging.Logger{
		ServiceName: ServiceName,
		Stage:       Settings.Stage,
		LoggerLevel: Settings.LoggerLevel,
	}
	Logger, Provider, SetupError := Configuration.CreateOTelLogger(os.Stdout)
	if SetupError != nil {
		return nil, SetupError
	}
	Lifecycle.Append(fx.Hook{OnStop: Provider.Shutdown})
	return Logger, nil
}

func LoggerProvider() fx.Option {
	return fx.Provide(CreateLogger)
}

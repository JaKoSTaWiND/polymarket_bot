package di_lifecycle

import (
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
)

func LifecycleProvider(GeneralState *states.GeneralState, Logger *slog.Logger) *lifecycle.Lifecycle {
	return &lifecycle.Lifecycle{GeneralState: GeneralState, Logger: Logger.With("component", "lifecycle", "bot_id", GeneralState.BotID)}
}

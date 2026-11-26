package actors

import (
	"context"
	"log/slog"

	usecases_lifecycle "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/lifecycle"
)

type UCS struct {
	StartBotUC *usecases_lifecycle.StartBotUC
}

type GeneralActor struct {
	Logger *slog.Logger
	UCS

}

func (GeneralActor *GeneralActor) StartBot(ctx context.Context) error {
	// sdk запрос на получение баланса (портфолио) (пока пропустим)

	return GeneralActor.UCS.StartBotUC.Execute(ctx)
}
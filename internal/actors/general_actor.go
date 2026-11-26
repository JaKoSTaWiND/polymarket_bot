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
	// sdk запрос на получение баланса +

	// подписывается на user channel +

	// получениие позиций +

	// создать акторов исходя из полученных позиций, причем если есть 2 позиции на 1 condition_id то для них создается 1 актор и принимает обе позиции в себя

	// если есть открытые позиции, то делает market channel subscribe на них
}
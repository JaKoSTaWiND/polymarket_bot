package actors_general

import (
	"context"
	"log/slog"

	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
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
	return nil
}

func (actor *GeneralActor) CreateMarketActor(ConditionID string) {
	state := actor.StartBotUC.GeneralState
	if _, exists := state.MarketActorsByConditionID[ConditionID]; exists {
		return
	}
	if len(state.MarketActorsByConditionID) >= state.MaxTrackingMarkets {
		return
	}

	state.MarketActorsByConditionID[ConditionID] = *actors_market.CreateMarketActor(actor.Logger)
}

func CreateMarketActors(...ConditionIDS string, ...AssetIDS string) 
	/* 
	for ConditionIDS {
		CreateMarketActor(ConditionID string)
	}
	*/

	// если MarketActorsByConditionID пустой то делает market channel subscribe с AssetIDS
	// если уже есть какие то маркет акторы то делает market channel request с AssetIDS

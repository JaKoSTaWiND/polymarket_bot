package di_actors

import (
	"context"
	"log/slog"
	"net/http"

	actors_general "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/general"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
)

// CreateGeneralActor получает из Fx состояние бота, HTTP-клиент и операции SDK для снимка аккаунта и двух WebSocket-каналов.
func CreateGeneralActor(
	logger *slog.Logger,
	state *states.GeneralState,
	botLifecycle *lifecycle.Lifecycle,
	dialer *websocket.Dialer,
	subscribe actors_general.MarketSubscribeFunc,
	update actors_general.MarketUpdateFunc,
	apis *states.ApisState,
	client *http.Client,
	balance actors_general.BalanceFunc,
	positions actors_general.OpenPositionsFunc,
	subscribeUser actors_general.UserSubscribeFunc,
	readUser actors_general.UserReadFunc,
) *actors_general.GeneralActor {
	return &actors_general.GeneralActor{
		Logger:       logger,
		GeneralState: state,
		Lifecycle:    botLifecycle,

		APIs:          apis,
		HTTPClient:    client,
		GetBalance:    balance,
		OpenPositions: positions,
		SubscribeUser: subscribeUser,
		ReadUser:      readUser,

		WSDialer:        dialer,
		SubscribeMarket: subscribe,
		UpdateMarket:    update,
	}
}

func registerGeneralActor(lifecycle fx.Lifecycle, actor *actors_general.GeneralActor, shutdowner fx.Shutdowner) {
	var cancel context.CancelFunc
	var stopped <-chan struct{}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			botCtx, stop := context.WithCancel(context.WithoutCancel(ctx))
			done := make(chan struct{})
			cancel = stop
			stopped = done

			go func() {
				defer close(done)
				if err := actor.StartBot(botCtx); err != nil && botCtx.Err() == nil {
					actor.Logger.ErrorContext(botCtx, "Бот завершился с ошибкой", "error", err)
					if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
						actor.Logger.ErrorContext(botCtx, "Не удалось инициировать остановку приложения", "error", err)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			defer actor.CloseMarketChannel()
			cancel()
			select {
			case <-stopped:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}

func GeneralActorProvider() fx.Option {
	return fx.Options(fx.Provide(CreateGeneralActor), fx.Invoke(registerGeneralActor))
}

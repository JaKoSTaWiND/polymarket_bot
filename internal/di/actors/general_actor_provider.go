package di_actors

import (
	"context"
	"log/slog"

	actors_general "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/general"
	usecases_lifecycle "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/lifecycle"
	"go.uber.org/fx"
)

func CreateGeneralActor(logger *slog.Logger, startBot *usecases_lifecycle.StartBotUC) *actors_general.GeneralActor {
	return &actors_general.GeneralActor{
		Logger: logger,
		UCS:    actors_general.UCS{StartBotUC: startBot},
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

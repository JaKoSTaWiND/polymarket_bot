package usecases_lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	usecases_sdk "github.com/JaKoSTaWiND/polymarket_bot/internal/usecases/sdk"
)

const (
	CHANGE_STATUS_ERR = "StartBotUC: Не удалось поменять статус"
)

type StartBotUC struct {
	Logger                 *slog.Logger
	GeneralState           *states.GeneralState
	Lifecycle              lifecycle.Lifecycle
	SubscribeUserChannelUC *usecases_sdk.SubscribeUserChannelUC
}

func (UC *StartBotUC) Execute(ctx context.Context) error {
	Logger := UC.Logger.With("operation", "StartBotUC.Execute")
	// поменять статус
	err := UC.Lifecycle.ChangeStatus(
		ctx,
		lifecycle.STARTED,
	)
	if err != nil {
		Logger.ErrorContext(
			ctx, CHANGE_STATUS_ERR,
			"error", err,
		)
		return fmt.Errorf(CHANGE_STATUS_ERR)
	}

	// вызвать UserChannelSubscribeUC
	return UC.SubscribeUserChannelUC.Execute(ctx, func(ctx context.Context, data []byte) error {
		var event struct {
			Type string `json:"event_type"`
			ID   string `json:"id"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return err
		}

		Logger.InfoContext(ctx, "Получено событие Polymarket", "type", event.Type, "id", event.ID)
		return nil
	})

}

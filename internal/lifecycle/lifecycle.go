package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
)

var (
	ENSURE_STATUS_ERR = "Lifecycle: Не удалось изменить статус с %s; Доступные статусы: %s"
	CHANGE_STATUS_ERR = "StartBotUC: Не удалось поменять статус"
)
type Lifecycle struct {
	Logger       *slog.Logger
	GeneralState *states.GeneralState
}

func (Lifecycle *Lifecycle) EnsureStatus(ctx context.Context, Allowed ...string) error {
	CurrentStatus := Lifecycle.GeneralState.BotStatus

	if len(Allowed) == 0 {
		return nil
	}

	for _, Status := range Allowed {
		if CurrentStatus == Status {
			Lifecycle.Logger.DebugContext(
				ctx, "Проверка статуса",
				"operation", "Lifecycle.EnsureStatus",
				"current_bot_status", CurrentStatus,
			)
			return nil
		}
	}
	AllowedNames := make([]string, 0, len(Allowed))
	for _, Status := range Allowed {
		AllowedNames = append(AllowedNames, Status)
	}
	return fmt.Errorf(ENSURE_STATUS_ERR, CurrentStatus, strings.Join(AllowedNames, ", "))
}

func (Lifecycle *Lifecycle) ChangeStatus(
	ctx context.Context,
	RequestStatus string,
	Allowed ...string,
) error {
	err := Lifecycle.EnsureStatus(ctx, Allowed...)
	if err != nil {
		// logging err
		return err
	}

	PreviousStatus := Lifecycle.GeneralState.BotStatus
	Lifecycle.GeneralState.BotStatus = RequestStatus
	Lifecycle.Logger.InfoContext(
		ctx, "Статус бота изменен",
		"operation", "Lifecycle.ChangeStatus",
		"previous_status", PreviousStatus,
		"new_status", RequestStatus,
	)
	return nil
}

func (Lifecycle *Lifecycle) StartBot(ctx context.Context) error {
	Logger := Lifecycle.Logger.With("operation", "Lifecycle.StartBot")
	// поменять статус
	err := Lifecycle.ChangeStatus(
		ctx,
		STARTED,
	)
	if err != nil {
		Logger.ErrorContext(
			ctx, CHANGE_STATUS_ERR,
			"error", err,
		)
		return fmt.Errorf(CHANGE_STATUS_ERR)
	}
	return nil
}
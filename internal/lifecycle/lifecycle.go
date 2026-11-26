package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
)

var (
	ENSURE_STATUS_ERROR = "Lifecycle: Не удалось изменить статус с %s; Доступные статусы: %s"
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
	return fmt.Errorf(ENSURE_STATUS_ERROR, CurrentStatus, strings.Join(AllowedNames, ", "))
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
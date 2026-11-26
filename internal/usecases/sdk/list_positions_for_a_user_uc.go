package usecases_sdk

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/sdk"
	sdk_profile "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/profile"
)

type ListPositionForAUserUC struct {
	Logger *slog.Logger
	Client *sdk.SDKClient
}

// Execute возвращает открытые и закрытые позиции кошелька walletAddress одним JSON-массивом.
// При ошибке SDK возвращает nil и ошибку. ctx передаётся в журнал ошибок.
func (UC *ListPositionForAUserUC) Execute(ctx context.Context, walletAddress string) (json.RawMessage, error) {
	positions, err := sdk_profile.ListPositionsForAUser(walletAddress)
	if err != nil && UC.Logger != nil {
		UC.Logger.ErrorContext(ctx, "Не удалось получить позиции кошелька",
			"operation", "ListPositionForAUserUC.Execute", "wallet", walletAddress, "error", err)
	}
	return positions, err
}

package actors_general

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"net/http"
	"slices"
	"strings"

	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
	sdk_market "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/market"
	sdkws "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/websocket"
	"github.com/gorilla/websocket"
)

// BalanceFunc читает через HTTP-клиент баланс кошелька в микро-pUSD.
type BalanceFunc = func(context.Context, *http.Client, string) (uint64, error)

// OpenPositionsFunc возвращает исходные строки Data API для открытых позиций кошелька.
type OpenPositionsFunc = func(context.Context, *http.Client, string) ([]json.RawMessage, error)

// UserSubscribeFunc открывает приватное соединение с переданными ключами API.
type UserSubscribeFunc = func(context.Context, *websocket.Dialer, sdkws.UserChannelSubscriptionRequest) (*websocket.Conn, error)

// UserReadFunc обслуживает соединение и закрывает его при выходе.
type UserReadFunc = func(context.Context, *websocket.Conn, func(context.Context, []byte) error) error

// startupPortfolio содержит позиции кошелька и адреса для общей рыночной подписки.
type startupPortfolio struct {
	byMarket    map[string]actors_market.PositionsByToken // Позиции кошелька по рынку и исходу.
	liveMarkets []string                                  // Торгуемые рынки, подтверждённые Gamma.
	liveTokens  []string                                  // Уникальные токены исходов этих рынков.
}

// groupOpenPositions группирует строки rows кошелька wallet и проверяет доступность торгов через client.
// ctx управляет запросами к Gamma. При ошибке разбора или HTTP портфель не публикуется.
//
//	Позиции -> группы по ConditionID -> статус рынка -> токены для подписки
func groupOpenPositions(ctx context.Context, client *http.Client, wallet string, rows []json.RawMessage) (startupPortfolio, error) {
	result := startupPortfolio{byMarket: make(map[string]actors_market.PositionsByToken)}
	candidates := make(map[string][]string)
	for i, raw := range rows {
		var row struct {
			Wallet      string      `json:"proxy_wallet"`      // Кошелёк позиции.
			Size        json.Number `json:"current_size"`      // Текущий остаток в токенах исхода.
			ConditionID string      `json:"condition_id"`      // Общий идентификатор рынка.
			TokenID     string      `json:"token_id"`          // Токен, которым владеет кошелёк.
			Opposite    string      `json:"opposite_token_id"` // Токен второго исхода, если известен.
			Status      string      `json:"status"`            // Состояние позиции, а не рынка.
			Redeemable  bool        `json:"redeemable"`        // Позицию уже можно погасить.
			Archived    bool        `json:"archived"`          // Архивный рынок.
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			return startupPortfolio{}, fmt.Errorf("позиция %d: %w", i, err)
		}
		if !strings.EqualFold(row.Wallet, wallet) || row.Status == "" || row.TokenID == "" || row.ConditionID == "" {
			return startupPortfolio{}, fmt.Errorf("позиция %d: некорректный кошелёк, статус или идентификатор", i)
		}
		if row.Status == "CLOSED" {
			continue
		}
		size, ok := new(big.Rat).SetString(row.Size.String())
		if !ok || size.Sign() < 0 {
			return startupPortfolio{}, fmt.Errorf("позиция %d: некорректный current_size", i)
		}
		if size.Sign() == 0 {
			continue
		}
		positions := result.byMarket[row.ConditionID]
		if positions == nil {
			positions = make(actors_market.PositionsByToken)
			result.byMarket[row.ConditionID] = positions
		}
		positions[row.TokenID] = raw
		if row.Redeemable || row.Archived || (row.Status != "OPEN" && row.Status != "MERGEABLE") {
			continue
		}
		candidates[row.ConditionID] = append(candidates[row.ConditionID], row.TokenID)
		if row.Opposite != "" {
			candidates[row.ConditionID] = append(candidates[row.ConditionID], row.Opposite)
		}
	}
	live, err := sdk_market.ListTradableMarkets(ctx, client, slices.Sorted(maps.Keys(candidates)))
	if err != nil {
		return startupPortfolio{}, fmt.Errorf("статусы рынков: %w", err)
	}
	tokens := make(map[string]struct{})
	for _, id := range live {
		for _, token := range candidates[id] {
			tokens[token] = struct{}{}
		}
	}
	result.liveMarkets = live
	result.liveTokens = slices.Sorted(maps.Keys(tokens))
	return result, nil
}

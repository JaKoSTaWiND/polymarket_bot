package sdk_market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
)

// ListTradableMarkets возвращает ConditionID открытых рынков, принимающих заявки, из conditionIDs.
// Запрашивает Gamma через client пакетами, ctx отменяет запрос. Пустой список обходится без сети.
func ListTradableMarkets(ctx context.Context, client *http.Client, conditionIDs []string) ([]string, error) {
	var live []string
	// Несколько десятков хешей помещаются в URL, а limit вмещает весь пакет.
	for batch := range slices.Chunk(conditionIDs, 50) {
		q := url.Values{"condition_ids": batch, "closed": {"false"}, "limit": {strconv.Itoa(len(batch))}}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gamma-api.polymarket.com/markets?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("Gamma /markets: HTTP %d", resp.StatusCode)
		}
		var markets []struct {
			ConditionID     string `json:"conditionId"`     // Ключ рынка в Gamma.
			Active          bool   `json:"active"`          // Рынок активен.
			Closed          *bool  `json:"closed"`          // Требуется явное подтверждение открытого рынка.
			Archived        bool   `json:"archived"`        // Архивные рынки доступны только для чтения.
			AcceptingOrders bool   `json:"acceptingOrders"` // Биржа принимает новые заявки.
			OrderBook       *bool  `json:"enableOrderBook"` // Явно отключённая книга заявок исключает рынок.
		}
		err = json.NewDecoder(resp.Body).Decode(&markets)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if markets == nil {
			return nil, fmt.Errorf("Gamma /markets: ожидался массив рынков")
		}
		for _, m := range markets {
			if !m.Active || m.Archived || m.Closed == nil || *m.Closed || !m.AcceptingOrders || (m.OrderBook != nil && !*m.OrderBook) {
				continue
			}
			if slices.Contains(batch, m.ConditionID) {
				live = append(live, m.ConditionID)
			}
		}
	}
	return live, nil
}

package sdk_profile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/transport/http"
)

// ListPositionsForAUser возвращает открытые и закрытые позиции торгового кошелька wallet.
// Результат — единый JSON-массив с исходными полями позиций. Пустой результат — [].
// Если запрос или разбор любой страницы завершается ошибкой, функция возвращает nil и ошибку.
//
//	OPEN (включая архивные рынки) -> CLOSED (только wallet) -> JSON
func ListPositionsForAUser(wallet string) (json.RawMessage, error) {
	positions, err := listPositions(context.Background(), transport_http.HttpClient(), wallet, "OPEN", "CLOSED")
	if err != nil {
		return nil, err
	}
	return json.Marshal(positions)
}

// listPositions загружает через client позиции wallet по выбранным statuses; ошибка отменяет весь результат.
func listPositions(ctx context.Context, client *http.Client, wallet string, statuses ...string) ([]json.RawMessage, error) {
	positions := make([]json.RawMessage, 0)
	for _, status := range statuses {
		query := url.Values{"user": {wallet}, "status": {status}, "limit": {"1000"}}
		if status == "OPEN" {
			query.Set("include_archived", "true")
			query.Set("filter_amount", "0")
		}
		seen := make(map[string]struct{})
		for {
			rows, next, err := getPositionsPage(ctx, client, query)
			if err != nil {
				return nil, fmt.Errorf("позиции %s: %w", status, err)
			}
			for _, row := range rows {
				// CLOSED может содержать позиции других кошельков того же аккаунта.
				if status == "CLOSED" {
					var owner struct {
						Wallet string `json:"proxy_wallet"` // Кошелёк, которому принадлежала позиция.
					}
					if err := json.Unmarshal(row, &owner); err != nil {
						return nil, err
					}
					if !strings.EqualFold(owner.Wallet, wallet) {
						continue
					}
				}
				positions = append(positions, row)
			}
			if next == "" {
				break
			}
			if _, repeated := seen[next]; repeated {
				return nil, fmt.Errorf("позиции %s: повтор курсора", status)
			}
			seen[next] = struct{}{}
			query.Set("cursor", next)
		}
	}
	return positions, nil
}

// getPositionsPage запрашивает через client страницу по query и возвращает позиции и следующий курсор.
// Пустой курсор завершает обход. Ошибка HTTP или формата ответа прерывает его.
func getPositionsPage(ctx context.Context, client *http.Client, query url.Values) ([]json.RawMessage, string, error) {
	const endpoint = "https://data-api.polymarket.com/v2/positions"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GET /v2/positions: %s", resp.Status)
	}

	var page struct {
		Data       []json.RawMessage `json:"data"` // Исходные JSON-объекты позиций.
		Pagination *struct {
			NextCursor string `json:"next_cursor"` // Пустая строка при JSON null.
			HasMore    bool   `json:"has_more"`    // Есть ли следующая страница.
		} `json:"pagination"` // Параметры продолжения обхода.
	}
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", err
	}
	if page.Data == nil || page.Pagination == nil || page.Pagination.HasMore != (page.Pagination.NextCursor != "") {
		return nil, "", fmt.Errorf("некорректная страница /v2/positions")
	}
	return page.Data, page.Pagination.NextCursor, nil
}

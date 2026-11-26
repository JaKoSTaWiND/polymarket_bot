package sdk_profile

import (
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
	client := transport_http.HttpClient()
	positions := make([]json.RawMessage, 0)
	for _, status := range []string{"OPEN", "CLOSED"} {
		query := url.Values{"user": {wallet}, "status": {status}, "limit": {"1000"}}
		if status == "OPEN" {
			query.Set("include_archived", "true")
			query.Set("filter_amount", "0")
		}
		for {
			rows, next, err := getPositionsPage(client, query)
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
			query.Set("cursor", next)
		}
	}
	return json.Marshal(positions)
}

// getPositionsPage запрашивает через client страницу по query и возвращает позиции и следующий курсор.
// Пустой курсор завершает обход. Ошибка HTTP или формата ответа прерывает его.
func getPositionsPage(client *http.Client, query url.Values) ([]json.RawMessage, string, error) {
	const endpoint = "https://data-api.polymarket.com/v2/positions"
	resp, err := client.Get(endpoint + "?" + query.Encode())
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

package sdk_markets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/entities"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/sdk"
	transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/transport/http"
	"github.com/shopspring/decimal"
)

// GetMarketByID загружает рынок из Gamma API по числовому идентификатору marketID.
// Денежные поля представлены как int64 в миллионных долях согласно docs/MONEY.md.
// Цены переводятся точно, а агрегаты Liquidity и Volume округляются до ближайшей миллионной.
// Отсутствующие поля получают нулевые значения.
// При неверном ID, сбое HTTP или некорректном ответе возвращает пустой рынок и ошибку.
//
//	[marketID] -> [GET /markets/{id}] -> [Разбор ответа] -> [entities.Market]
//	                      |                     |
//	                      +-------- error ------+
func GetMarketByID(marketID string) (entities.Market, error) {
	if _, err := strconv.ParseUint(marketID, 10, 64); err != nil {
		return entities.Market{}, fmt.Errorf("некорректный ID рынка %q: %w", marketID, err)
	}

	path := "/markets/" + marketID
	resp, err := transport_http.HttpClient().Get(sdk.GAMMA_URL + path)
	if err != nil {
		return entities.Market{}, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return entities.Market{}, fmt.Errorf("GET %s: HTTP %d", path, resp.StatusCode)
	}

	var reply struct {
		ID           string     `json:"id"`            // Числовой ID рынка в представлении Gamma.
		Slug         string     `json:"slug"`          // Часть URL страницы рынка.
		ConditionID  string     `json:"conditionId"`   // Идентификатор условия расчёта.
		Question     string     `json:"question"`      // Вопрос рынка.
		EndDate      *time.Time `json:"endDate"`       // Плановая дата окончания, если задана.
		StartDate    *time.Time `json:"startDate"`     // Дата начала, если задана.
		Liquidity    string     `json:"liquidity"`     // Десятичное значение ликвидности.
		Volume       string     `json:"volume"`        // Десятичное значение объёма.
		Outcomes     string     `json:"outcomes"`      // JSON-массив названий исходов внутри строки.
		Prices       string     `json:"outcomePrices"` // JSON-массив десятичных цен внутри строки.
		Version      string     `json:"version"`       // Версия протокола расчёта.
		ClobTokenIDs string     `json:"clobTokenIds"`  // JSON-массив ID токенов CTF внутри строки.
		PositionIDs  []string   `json:"positionIds"`   // ID позиций Polymarket V2.
	}
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return entities.Market{}, fmt.Errorf("ответ Gamma для %s: %w", path, err)
	}
	if reply.ID == "" {
		return entities.Market{}, fmt.Errorf("ответ Gamma для %s: отсутствует id", path)
	}

	// У агрегированных метрик Gamma может быть больше шести дробных знаков.
	liquidity, err := marketMicroUnits(reply.Liquidity, true)
	if err != nil {
		return entities.Market{}, fmt.Errorf("%s, liquidity: %w", path, err)
	}
	volume, err := marketMicroUnits(reply.Volume, true)
	if err != nil {
		return entities.Market{}, fmt.Errorf("%s, volume: %w", path, err)
	}

	outcomes, err := decodeMarketPair(reply.Outcomes)
	if err != nil {
		return entities.Market{}, fmt.Errorf("%s, outcomes: %w", path, err)
	}
	prices, err := decodeMarketPair(reply.Prices)
	if err != nil {
		return entities.Market{}, fmt.Errorf("%s, outcomePrices: %w", path, err)
	}

	// Версия определяет источник AssetID, даже когда Gamma заполняет оба поля.
	var assetIDs [2]string
	switch reply.Version {
	case "v2":
		if len(reply.PositionIDs) != 0 && len(reply.PositionIDs) != len(assetIDs) {
			return entities.Market{}, fmt.Errorf("%s, positionIds: ожидалось два идентификатора", path)
		}
		copy(assetIDs[:], reply.PositionIDs)
	case "", "v1":
		assetIDs, err = decodeMarketPair(reply.ClobTokenIDs)
		if err != nil {
			return entities.Market{}, fmt.Errorf("%s, clobTokenIds: %w", path, err)
		}
	default:
		return entities.Market{}, fmt.Errorf("%s: неизвестная версия рынка %q", path, reply.Version)
	}

	// Название, цена и идентификатор одного исхода имеют одинаковый индекс в ответе.
	var tokens [2]entities.Token
	for i := range tokens {
		price, err := marketMicroUnits(prices[i], false)
		if err != nil {
			return entities.Market{}, fmt.Errorf("%s, outcomePrices[%d]: %w", path, i, err)
		}
		tokens[i] = entities.Token{AssetID: assetIDs[i], Outcome: outcomes[i], Price: price}
	}

	return entities.Market{
		MarketID:       reply.ID,
		Slug:           reply.Slug,
		ConditionID:    reply.ConditionID,
		Question:       reply.Question,
		EndDate:        reply.EndDate,
		StartDate:      reply.StartDate,
		Liquidity:      liquidity,
		Volume:         volume,
		PrimaryToken:   tokens[0],
		SecondaryToken: tokens[1],
	}, nil
}

// decodeMarketPair разбирает JSON-массив encoded из двух значений в порядке Gamma.
// Пустое значение даёт пару пустых строк. Некорректный JSON или длина дают ошибку.
func decodeMarketPair(encoded string) ([2]string, error) {
	var pair [2]string
	if encoded == "" {
		return pair, nil
	}

	var values []string
	if err := json.Unmarshal([]byte(encoded), &values); err != nil {
		return pair, err
	}
	if len(values) != 0 && len(values) != len(pair) {
		return pair, fmt.Errorf("ожидалось два значения, получено %d", len(values))
	}
	copy(pair[:], values)
	return pair, nil
}

// marketMicroUnits переводит десятичную строку value в миллионные доли.
// round разрешает округление для агрегированных метрик. Пустая строка даёт 0.
// Неверный формат, отрицательное значение, потеря точности при round == false
// и переполнение дают ошибку.
func marketMicroUnits(value string, round bool) (int64, error) {
	if value == "" {
		return 0, nil
	}
	number, err := decimal.NewFromString(value)
	if err != nil {
		return 0, err
	}
	if number.IsNegative() {
		return 0, fmt.Errorf("ожидалось неотрицательное значение, получено %q", value)
	}

	scaled := number.Shift(6)
	if round {
		scaled = scaled.Round(0)
	} else if !scaled.IsInteger() {
		return 0, fmt.Errorf("значение %q нельзя точно выразить в миллионных долях", value)
	}

	units := scaled.BigInt()
	if !units.IsInt64() {
		return 0, fmt.Errorf("значение %q не помещается в int64 после пересчёта", value)
	}
	return units.Int64(), nil
}

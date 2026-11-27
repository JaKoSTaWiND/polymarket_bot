package sdk_websocket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
)

const marketChannelURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

// MarketChannelSubscriptionRequest задаёт начальную подписку на токены.
type MarketChannelSubscriptionRequest struct {
	AssetIDS             []string `json:"assets_ids"`                       // Идентификаторы токенов исходов.
	Type                 string   `json:"type"`                             // Пустое значение при подписке заменяется на market.
	InitialDump          *bool    `json:"initial_dump,omitempty"`           // nil включает начальный снимок по умолчанию.
	Level                int      `json:"level,omitempty"`                  // Уровень от 1 до 3. При 0 сервер использует 2.
	CustomFeatureEnabled bool     `json:"custom_feature_enabled,omitempty"` // Включает best_bid_ask, new_market и market_resolved.
}

// MarketChannelUpdateRequest меняет набор токенов на открытом соединении.
type MarketChannelUpdateRequest struct {
	Operation            string   `json:"operation"`                        // subscribe или unsubscribe.
	AssetIDS             []string `json:"assets_ids"`                       // Токены для добавления или удаления.
	Level                int      `json:"level,omitempty"`                  // Уровень от 1 до 3. При 0 поле опускается.
	CustomFeatureEnabled *bool    `json:"custom_feature_enabled,omitempty"` // nil пропускает поле, false передаётся явно.
}

// MarketChannelSubscribe открывает публичный канал и отправляет начальную подписку req.
// ctx управляет подключением через dialer. При nil используются настройки Gorilla по умолчанию.
// Вызывающий код закрывает соединение и отправляет текстовый PING каждые 10 секунд.
//
//	Проверка req -> подключение -> запись JSON -> открытое соединение
//	                                   |
//	                            ошибка записи -> закрытие
func MarketChannelSubscribe(ctx context.Context, dialer *websocket.Dialer, req MarketChannelSubscriptionRequest) (*websocket.Conn, error) {
	if req.Type != "" && req.Type != "market" {
		return nil, fmt.Errorf("недопустимый тип канала %q", req.Type)
	}
	if err := validateMarketSubscription(req.AssetIDS, req.Level); err != nil {
		return nil, err
	}
	req.Type = "market"

	conn, _, err := dialer.DialContext(ctx, marketChannelURL, nil)
	if err != nil {
		return nil, fmt.Errorf("подключение к Market Channel: %w", err)
	}
	if err := conn.WriteJSON(req); err != nil {
		conn.Close()
		return nil, fmt.Errorf("начальная подписка Market Channel: %w", err)
	}
	return conn, nil
}

// MarketChannelUpdate отправляет req по открытому conn и возвращает ошибку проверки или записи.
// Владелец соединения сериализует этот вызов с другими записями, включая PING.
func MarketChannelUpdate(conn *websocket.Conn, req MarketChannelUpdateRequest) error {
	if conn == nil {
		return errors.New("отсутствует соединение с Market Channel")
	}
	if req.Operation != "subscribe" && req.Operation != "unsubscribe" {
		return fmt.Errorf("недопустимая операция подписки %q", req.Operation)
	}
	if err := validateMarketSubscription(req.AssetIDS, req.Level); err != nil {
		return err
	}
	if err := conn.WriteJSON(req); err != nil {
		return fmt.Errorf("обновление подписки Market Channel: %w", err)
	}
	return nil
}

// validateMarketSubscription проверяет список токенов assetIDs и допустимый уровень level до отправки.
func validateMarketSubscription(assetIDs []string, level int) error {
	if len(assetIDs) == 0 {
		return errors.New("для подписки нужен хотя бы один идентификатор токена")
	}
	for i, id := range assetIDs {
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("пустой идентификатор в assets_ids[%d]", i)
		}
	}
	if level < 0 || level > 3 {
		return fmt.Errorf("недопустимый уровень подписки %d", level)
	}
	return nil
}

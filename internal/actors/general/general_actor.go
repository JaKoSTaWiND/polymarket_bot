package actors_general

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	sdkws "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/websocket"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"
)

// BalanceFunc читает через HTTP-клиент баланс кошелька в микро-pUSD.
type BalanceFunc = func(context.Context, *http.Client, string) (uint64, error)

// UserSubscribeFunc открывает приватное соединение с переданными ключами API.
type UserSubscribeFunc = func(context.Context, *websocket.Dialer, sdkws.UserChannelSubscriptionRequest) (*websocket.Conn, error)

// UserReadFunc обслуживает соединение и закрывает его при выходе.
type UserReadFunc = func(context.Context, *websocket.Conn, func(context.Context, []byte) error) error

// MarketSubscribeFunc открывает Market Channel через dialer и отправляет начальную подписку.
type MarketSubscribeFunc = func(context.Context, *websocket.Dialer, sdkws.MarketChannelSubscriptionRequest) (*websocket.Conn, error)

// MarketUpdateFunc меняет подписку на уже открытом соединении.
type MarketUpdateFunc = func(*websocket.Conn, sdkws.MarketChannelUpdateRequest) error

// GeneralActor запускает бота и управляет реестром акторов рынков.
type GeneralActor struct {
	Logger       *slog.Logger         // Журнал актора и создаваемых им акторов рынков.
	GeneralState *states.GeneralState // Общее состояние с реестром и лимитом акторов.
	Lifecycle    *lifecycle.Lifecycle // Управляет статусом бота при запуске.

	APIs          *states.ApisState // Кошелёк и ключи для приватной подписки.
	HTTPClient    *http.Client      // Транспорт для запроса баланса.
	GetBalance    BalanceFunc       // Баланс в микро-pUSD.
	SubscribeUser UserSubscribeFunc // Открывает приватный канал аккаунта.
	ReadUser      UserReadFunc      // Читает приватный канал до отмены контекста.

	WSDialer        *websocket.Dialer   // Настройки подключения из DI.
	SubscribeMarket MarketSubscribeFunc // Открывает общую подписку.
	UpdateMarket    MarketUpdateFunc    // Добавляет токены в текущую подписку.

	startupMu         sync.Mutex          // Ограничивает StartBot одним активным вызовом.
	marketMu          sync.Mutex          // Сериализует подписки, отправку PING и закрытие соединения.
	marketConn        *websocket.Conn     // Общее соединение для всех отслеживаемых рынков.
	subscribedMarkets map[string]struct{} // ConditionID из успешно отправленных подписок текущего соединения.
}

// StartBot получает баланс кошелька и открывает приватную подписку User Channel.
// После успешной подписки публикует баланс и читает канал до отмены ctx или ошибки чтения.
// ctx управляет запросом баланса, подключением и чтением. Второй активный вызов возвращает ошибку.
//
//	Запрос баланса -> подписка User Channel -> запись баланса -> чтение канала
func (actor *GeneralActor) StartBot(ctx context.Context) error {
	if !actor.startupMu.TryLock() {
		return errors.New("StartBot уже выполняется")
	}
	defer actor.startupMu.Unlock()

	wallet := actor.APIs.PolyWalletAddress
	auth := sdkws.UserChannelAuth{ApiKey: actor.APIs.PolyApiKey, Secret: actor.APIs.PolySecret, Passphrase: actor.APIs.PolyPassphrase}
	if auth.ApiKey == "" || auth.Secret == "" || auth.Passphrase == "" {
		return errors.New("для User Channel нужны API key, secret и passphrase")
	}
	balance, err := actor.GetBalance(ctx, actor.HTTPClient, wallet)
	if err != nil {
		return fmt.Errorf("баланс кошелька: %w", err)
	}

	conn, err := actor.SubscribeUser(ctx, actor.WSDialer, sdkws.UserChannelSubscriptionRequest{Auth: auth, Type: "user"})
	if err != nil {
		return fmt.Errorf("подписка User Channel: %w", err)
	}

	atomic.StoreUint64(&actor.GeneralState.Balance, balance)
	actor.Logger.InfoContext(ctx, "Получен стартовый баланс", "balance_micro_pusd", balance)

	// ReadUser поддерживает соединение и закрывает его перед возвратом.
	err = actor.ReadUser(ctx, conn, func(ctx context.Context, payload []byte) error {
		actor.Logger.DebugContext(ctx, "Получено сообщение User Channel", "bytes", len(payload))
		return nil
	})
	if cause := context.Cause(ctx); cause != nil {
		return cause
	}
	if err == nil {
		err = io.EOF
	}
	return fmt.Errorf("чтение User Channel: %w", err)
}

func (actor *GeneralActor) CreateMarketActor(ConditionID string) {
	state := actor.GeneralState
	if _, exists := state.MarketActorsByConditionID[ConditionID]; exists {
		return
	}
	if len(state.MarketActorsByConditionID) >= state.MaxTrackingMarkets {
		return
	}

	state.MarketActorsByConditionID[ConditionID] = *actors_market.CreateMarketActor(actor.Logger)
}

// CreateMarketActors создаёт акторов для новых рынков из ConditionIDS.
// Если свободных мест недостаточно для всех уникальных новых ID, метод отклоняет всю пачку до создания первого актора.
func (actor *GeneralActor) CreateMarketActors(ConditionIDS []string) {
	state := actor.GeneralState
	freeSlots := state.MaxTrackingMarkets - len(state.MarketActorsByConditionID)
	newIDs := make(map[string]struct{})

	for _, conditionID := range ConditionIDS {
		if _, exists := state.MarketActorsByConditionID[conditionID]; exists {
			continue
		}
		newIDs[conditionID] = struct{}{}
		if len(newIDs) > freeSlots {
			return
		}
	}

	for _, conditionID := range ConditionIDS {
		actor.CreateMarketActor(conditionID)
	}
}

// KillMarketActor удаляет актора рынка ConditionID из реестра и освобождает его место.
func (actor *GeneralActor) KillMarketActor(ConditionID string) {
	delete(actor.GeneralState.MarketActorsByConditionID, ConditionID)
}

// KillMarketActors удаляет из реестра акторов для ConditionIDS. Повторы и неизвестные ID пропускаются.
func (actor *GeneralActor) KillMarketActors(ConditionIDS []string) {
	for _, conditionID := range ConditionIDS {
		actor.KillMarketActor(conditionID)
	}
}

// func StartTrackMarket()

// StartTrackMarkets создаёт по одному актору на ConditionID и подписывается на все AssetIDS.
// Один ConditionID может соответствовать нескольким токенам. Ошибка сохраняет прежний состав реестра.
//
//	Проверка вместимости -> создание акторов -> подписка или обновление -> при ошибке откат
func (actor *GeneralActor) StartTrackMarkets(ConditionIDS []string, AssetIDS []string) error {
	if len(ConditionIDS) == 0 {
		return errors.New("для отслеживания нужен хотя бы один ConditionID")
	}
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	registry := actor.GeneralState.MarketActorsByConditionID
	countBefore := len(registry)
	var newIDs []string
	for _, id := range ConditionIDS {
		if _, exists := registry[id]; !exists {
			newIDs = append(newIDs, id)
		}
	}

	// CreateMarketActors принимает или отклоняет пачку целиком.
	actor.CreateMarketActors(newIDs)
	if len(newIDs) > 0 && len(registry) == countBefore {
		return errors.New("недостаточно мест для всей пачки рынков")
	}

	var err error
	if countBefore == 0 {
		// Удаление последнего актора могло оставить старое соединение открытым.
		if actor.marketConn != nil {
			actor.marketConn.Close()
		}
		req := sdkws.MarketChannelSubscriptionRequest{AssetIDS: AssetIDS}
		actor.subscribedMarkets = nil
		actor.marketConn, err = actor.SubscribeMarket(context.Background(), actor.WSDialer, req)
		if err == nil {
			actor.subscribedMarkets = make(map[string]struct{})
			go actor.keepMarketChannelAlive(actor.marketConn)
		}
	} else if actor.marketConn == nil {
		err = errors.New("отсутствует соединение Market Channel для обновления подписки")
	} else if err = actor.marketConn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err == nil {
		req := sdkws.MarketChannelUpdateRequest{Operation: "subscribe", AssetIDS: AssetIDS}
		err = actor.UpdateMarket(actor.marketConn, req)
	}
	if err != nil {
		actor.KillMarketActors(newIDs)
		return err
	}

	for _, id := range ConditionIDS {
		actor.subscribedMarkets[id] = struct{}{}
	}
	if actor.Logger != nil {
		ids := slices.Sorted(maps.Keys(actor.subscribedMarkets))
		actor.Logger.Info("Отправлена подписка Market Channel", "condition_ids", attribute.StringSliceValue(ids))
	}
	return nil
}

// OnMessage записывает в журнал исходный текст сообщения Market Channel из payload.
func (actor *GeneralActor) OnMessage(payload []byte) {
	if actor.Logger != nil {
		actor.Logger.Info("Получено сообщение Market Channel", "payload", string(payload))
	}
}

// keepMarketChannelAlive обслуживает conn до закрытия. Каждое прочитанное сообщение передаётся в OnMessage.
//
//	Чтение -> закрытие канала done -> остановка таймера и соединения
func (actor *GeneralActor) keepMarketChannelAlive(conn *websocket.Conn) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			actor.OnMessage(payload)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer func() {
		conn.Close()
		actor.marketMu.Lock()
		defer actor.marketMu.Unlock()
		if actor.marketConn == conn {
			actor.marketConn = nil
			actor.subscribedMarkets = nil
		}
	}()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			actor.marketMu.Lock()
			err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, []byte("PING"))
			}
			actor.marketMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// CloseMarketChannel освобождает общее соединение при остановке бота.
func (actor *GeneralActor) CloseMarketChannel() {
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()
	if actor.marketConn != nil {
		actor.marketConn.Close()
		actor.marketConn = nil
		actor.subscribedMarkets = nil
	}
}

// func StopTrackMarket()
// func StopTrackMarkets()

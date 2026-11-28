package actors_general

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/lifecycle"
	sdk_markets "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/markets"
	sdkws "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/websocket"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
	"github.com/gorilla/websocket"
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
	marketMu          sync.Mutex          // Защищает реестр акторов и общее соединение Market Channel.
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
	actor.Logger.InfoContext(ctx, "Получен стартовый баланс", "balance", balance)

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

// CreateMarketActor загружает рынок по числовому marketID из Gamma и возвращает актор из реестра.
// Для нового ConditionID регистрирует актор под marketMu. Существующий актор сохраняет своё состояние.
// Сбой загрузки, пустой ConditionID и исчерпание лимита для нового рынка возвращают ошибку.
//
//	[Gamma] -> [Проверка ConditionID и лимита] -> [Получение или регистрация актора]
func (actor *GeneralActor) CreateMarketActor(marketID string) (*actors_market.MarketActor, error) {
	market, err := sdk_markets.GetMarketByID(marketID)
	if err != nil {
		return nil, err
	}
	if market.ConditionID == "" {
		return nil, fmt.Errorf("рынок %q не содержит ConditionID", marketID)
	}

	// Проверка лимита и запись выполняются под одной блокировкой.
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	state := actor.GeneralState
	if marketActor, exists := state.MarketActorsByConditionID[market.ConditionID]; exists {
		return marketActor, nil
	}
	if len(state.MarketActorsByConditionID) >= state.MaxTrackingMarkets {
		return nil, fmt.Errorf("достигнут лимит акторов рынков (%d)", state.MaxTrackingMarkets)
	}

	marketActor := actors_market.CreateMarketActor(actor.Logger, market)
	state.MarketActorsByConditionID[market.ConditionID] = marketActor
	return marketActor, nil
}

// KillMarketActor удаляет актора по conditionID и освобождает его место.
func (actor *GeneralActor) KillMarketActor(conditionID string) {
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	delete(actor.GeneralState.MarketActorsByConditionID, conditionID)
}

// StartTrackMarket подписывает оба исхода рынка с числовым marketID из Gamma.
// ctx управляет открытием WebSocket. Ошибка создания актора завершает вызов до подписки.
// При сбое WebSocket актор остаётся в реестре для повторной попытки.
//
//	[CreateMarketActor] -> [Проверка реестра и подписки]
//	                       +-- подписка есть -> возврат nil
//	                       +-- соединения нет -> SubscribeMarket
//	                       +-- соединение есть -> UpdateMarket
func (actor *GeneralActor) StartTrackMarket(ctx context.Context, marketID string) error {
	marketActor, err := actor.CreateMarketActor(marketID)
	if err != nil {
		return err
	}

	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	market := marketActor.Market
	// KillMarketActor мог удалить актор до захвата блокировки.
	if actor.GeneralState.MarketActorsByConditionID[market.ConditionID] != marketActor {
		return fmt.Errorf("актор рынка %q удалён перед подпиской", marketID)
	}
	if _, subscribed := actor.subscribedMarkets[market.ConditionID]; subscribed {
		return nil
	}
	assetIDs := []string{market.PrimaryToken.AssetID, market.SecondaryToken.AssetID}

	// После разрыва соединения реестр может содержать акторы без подписки.
	if actor.marketConn == nil {
		conn, err := actor.SubscribeMarket(ctx, actor.WSDialer, sdkws.MarketChannelSubscriptionRequest{
			AssetIDS: assetIDs,
			Type:     "market",
		})
		if err != nil {
			return fmt.Errorf("подписка Market Channel для рынка %q: %w", marketID, err)
		}

		actor.marketConn = conn
		actor.subscribedMarkets = map[string]struct{}{market.ConditionID: {}}
		go actor.keepMarketChannelAlive(conn)
		return nil
	}

	// У обновления свой дедлайн: предыдущий дедлайн PING мог истечь.
	if err := actor.marketConn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("дедлайн записи Market Channel: %w", err)
	}
	if err := actor.UpdateMarket(actor.marketConn, sdkws.MarketChannelUpdateRequest{
		Operation: "subscribe",
		AssetIDS:  assetIDs,
	}); err != nil {
		return fmt.Errorf("добавление рынка %q в Market Channel: %w", marketID, err)
	}

	actor.subscribedMarkets[market.ConditionID] = struct{}{}
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

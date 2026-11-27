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

// CreateMarketActor загружает рынок по числовому marketID из Gamma и регистрирует актор по ConditionID.
// Существующий актор сохраняет своё состояние. Сбой загрузки, пустой ConditionID и исчерпание лимита
// для нового рынка возвращают ошибку.
//
//	Загрузка рынка -> проверка ConditionID и лимита -> регистрация актора
func (actor *GeneralActor) CreateMarketActor(marketID string) error {
	market, err := sdk_markets.GetMarketByID(marketID)
	if err != nil {
		return err
	}
	if market.ConditionID == "" {
		return fmt.Errorf("рынок %q не содержит ConditionID", marketID)
	}

	// Проверка лимита и запись выполняются под одной блокировкой.
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	state := actor.GeneralState
	if _, exists := state.MarketActorsByConditionID[market.ConditionID]; exists {
		return nil
	}
	if len(state.MarketActorsByConditionID) >= state.MaxTrackingMarkets {
		return fmt.Errorf("достигнут лимит акторов рынков (%d)", state.MaxTrackingMarkets)
	}

	state.MarketActorsByConditionID[market.ConditionID] = actors_market.CreateMarketActor(actor.Logger, market)
	return nil
}

// KillMarketActor удаляет актора по conditionID и освобождает его место.
func (actor *GeneralActor) KillMarketActor(conditionID string) {
	actor.marketMu.Lock()
	defer actor.marketMu.Unlock()

	delete(actor.GeneralState.MarketActorsByConditionID, conditionID)
}

// func StartTrackMarket()

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

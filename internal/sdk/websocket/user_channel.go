package sdk_websocket

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const URL = "wss://ws-subscriptions-clob.polymarket.com/ws/user"

type UserChannelAuth struct {
	ApiKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

type UserChannelSubscriptionRequest struct {
	Auth    UserChannelAuth     `json:"auth"`
	Type    string   `json:"type"`
	Markets []string `json:"markets,omitempty"`
}

func UserChannelSubscribe(
	ctx context.Context,
	Dialer *websocket.Dialer,
	UserChannelSubscriptionRequest UserChannelSubscriptionRequest,
) (*websocket.Conn, error) {
	Conn, _, err := Dialer.DialContext(ctx, URL, nil)
	if err != nil {
		return nil, err
	}
	if err := Conn.WriteJSON(UserChannelSubscriptionRequest); err != nil {
		Conn.Close()
		return nil, err
	}
	return Conn, nil
}

func ReadUserChannel(ctx context.Context, conn *websocket.Conn, onEvent func(context.Context, []byte) error) error {
	streamCtx, cancel := context.WithCancelCause(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer conn.Close()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-streamCtx.Done():
				return
			case <-ticker.C:
				err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err == nil {
					err = conn.WriteMessage(websocket.TextMessage, []byte("PING"))
				}
				if err != nil {
					cancel(fmt.Errorf("отправка PING: %w", err))
					return
				}
			}
		}
	}()
	defer func() {
		cancel(nil)
		<-done
	}()

	for {
		_, payload, err := conn.ReadMessage()
		if cause := context.Cause(streamCtx); cause != nil {
			return cause
		}
		if err != nil {
			return err
		}
		if string(payload) == "PONG" {
			continue
		}
		if err := onEvent(streamCtx, payload); err != nil {
			return err
		}
	}
}

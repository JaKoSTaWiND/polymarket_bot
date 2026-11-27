package di_sdk

import (
	sdk_websocket "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/websocket"
	"go.uber.org/fx"
)

func WebSocketProvider() fx.Option {
	return fx.Supply(
		sdk_websocket.UserChannelSubscribe,
		sdk_websocket.ReadUserChannel,
	)
}

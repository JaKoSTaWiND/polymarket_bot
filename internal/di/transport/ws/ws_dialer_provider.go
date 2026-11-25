package di_transport_ws

import (
	transport_ws "github.com/JaKoSTaWiND/polymarket_bot/internal/transport/ws"
	"go.uber.org/fx"
)

func WSDialerProvider() fx.Option {
	return fx.Provide(transport_ws.WSDialer)
}

package di_transport_http

import (
	transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/transport/http"
	"go.uber.org/fx"
)

func HttpClientProvider() fx.Option {
	return fx.Provide(transport_http.HttpClient)
}

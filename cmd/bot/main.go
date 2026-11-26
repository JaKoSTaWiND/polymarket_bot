package main

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/di"
	di_observability "github.com/JaKoSTaWiND/polymarket_bot/internal/di/observability"
	di_sdk "github.com/JaKoSTaWiND/polymarket_bot/internal/di/sdk"
	di_states "github.com/JaKoSTaWiND/polymarket_bot/internal/di/states"
	di_transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/di/transport/http"
	di_transport_ws "github.com/JaKoSTaWiND/polymarket_bot/internal/di/transport/ws"
	"go.uber.org/fx"
)

func Main() {
	BoostrapSettings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&BoostrapSettings), // Boostrap Settings

		// OBSERVABILITY
		di_observability.LoggerProvider(), // Logger

		// TRANSPORT
		di_transport_ws.WSDialerProvider(),     // Websocker
		di_transport_http.HttpClientProvider(), // HTTP Client
		//место под http server

		// STATES
		di_states.GeneralStateProvider(),

		// SDK
		di_sdk.SDKClientProvider(), // SDK Client
	)

	App.Run()
}

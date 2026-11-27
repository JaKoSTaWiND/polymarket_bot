package main

import (
	"log"

	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/di"
	di_actors "github.com/JaKoSTaWiND/polymarket_bot/internal/di/actors"
	di_lifecycle "github.com/JaKoSTaWiND/polymarket_bot/internal/di/lifecycle"
	di_observability "github.com/JaKoSTaWiND/polymarket_bot/internal/di/observability"
	di_sdk "github.com/JaKoSTaWiND/polymarket_bot/internal/di/sdk"
	di_states "github.com/JaKoSTaWiND/polymarket_bot/internal/di/states"
	di_transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/di/transport/http"
	di_transport_ws "github.com/JaKoSTaWiND/polymarket_bot/internal/di/transport/ws"
	"go.uber.org/fx"
)

func main() {
	BoostrapSettings, err := internal.LoadBootstrapSettings()
	if err != nil {
		log.Fatal(err)
	}
	App := fx.New(
		di.BootstrapSettingsProvider(&BoostrapSettings), // Boostrap Settings

		// OBSERVABILITY
		di_observability.LoggerProvider(), // Logger

		// TRANSPORT
		di_transport_ws.WSDialerProvider(),     // Websocker
		di_transport_http.HttpClientProvider(), // HTTP Client
		//место под http server

		// STATES
		di_states.GeneralStateProvider(), // General State
		di_states.ApisStateProvider(),    // APIs State

		// SDK
		di_sdk.SDKClientProvider(), // SDK Client
		di_sdk.ProfileProvider(),
		di_sdk.WebSocketProvider(),
		di_sdk.OtherProvider(),

		// LIFECYCLE
		fx.Provide(di_lifecycle.LifecycleProvider),

		// ACTORS
		di_actors.GeneralActorProvider(), // General Actor
	)

	App.Run()


	// тут будет запуск бота
	// Lifecycle.StartBot()
	// GeneralActor.StartBot()

}

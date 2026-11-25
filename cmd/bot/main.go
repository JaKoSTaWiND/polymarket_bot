package main

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/di"
	di_transport_ws "github.com/JaKoSTaWiND/polymarket_bot/internal/di/transport/ws"
	"go.uber.org/fx"
)

func Main() {
	BoostrapSettings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&BoostrapSettings),
		di_transport_ws.WSDialerProvider(),
	)

	App.Run()
}

package main

import (
	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/di"
	"go.uber.org/fx"
)

func Main() {
	BoostrapSettings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&BoostrapSettings),
	)

	App.Run()
}

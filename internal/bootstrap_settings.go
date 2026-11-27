package internal

import (
	"os"
	"strconv"
)

type BootstrapSettings struct {
	Stage                    string // Окружение приложения из STAGE.
	LoggerLevel              string // Минимальный уровень OTel Logs из LOGGER_LEVEL (DEBUG или INFO).
	PolyWalletAddress        string
	PolyRelayerApiKey        string
	PolyRelayerApiKeyAddress string
	PolyApiKey               string
	PolySecret               string
	PolyPassphrase           string
	MaxTrackingMarkets       int // Максимальное число одновременно отслеживаемых рынков.
}

// LoadBootstrapSettings читает настройки запуска из окружения.
// Пустой MAX_TRACKING_MARKETS даёт нулевой лимит. Значение, которое нельзя
// представить неотрицательным int, приводит к ошибке загрузки настроек.
func LoadBootstrapSettings() (BootstrapSettings, error) {
	MaxTrackingMarkets, Err := strconv.Atoi(os.Getenv("MAX_TRACKING_MARKETS"))
	if Err != nil {
		return BootstrapSettings{}, Err
	}

	return BootstrapSettings{
		Stage:                    os.Getenv("STAGE"),
		LoggerLevel:              os.Getenv("LOGGER_LEVEL"),
		PolyWalletAddress:        os.Getenv("POLY_WALLET_ADDRESS"),
		PolyRelayerApiKey:        os.Getenv("POLY_RELAYER_API_KEY"),
		PolyRelayerApiKeyAddress: os.Getenv("POLY_RELAYER_API_KEY_ADDRESS"),
		PolyApiKey:               os.Getenv("POLY_API_KEY"),
		PolySecret:               os.Getenv("POLY_SECRET"),
		PolyPassphrase:           os.Getenv("POLY_PASSPHRASE"),
		MaxTrackingMarkets:       MaxTrackingMarkets,
	}, nil
}

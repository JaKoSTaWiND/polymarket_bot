package internal

import "os"

type BootstrapSettings struct {
	Stage                    string // Окружение приложения из STAGE.
	LoggerLevel              string // Минимальный уровень OTel Logs из LOGGER_LEVEL (DEBUG или INFO).
	PolyRelayerApiKey        string
	PolyRelayerApiKeyAddress string
	PolyApiKey               string
	PolySecret               string
	PolyPassphrase           string
}

// LoadBootstrapSettings возвращает настройки из окружения без проверки значений.
// Отсутствующие строковые настройки остаются пустыми; секреты не логируются.
func LoadBootstrapSettings() BootstrapSettings {
	return BootstrapSettings{
		Stage:                    os.Getenv("STAGE"),
		LoggerLevel:              os.Getenv("LOGGER_LEVEL"),
		PolyRelayerApiKey:        os.Getenv("POLY_RELAYER_API_KEY"),
		PolyRelayerApiKeyAddress: os.Getenv("POLY_RELAYER_API_KEY_ADDRESS"),
		PolyApiKey:               os.Getenv("POLY_API_KEY"),
		PolySecret:               os.Getenv("POLY_SECRET"),
		PolyPassphrase:           os.Getenv("POLY_PASSPHRASE"),
	}
}

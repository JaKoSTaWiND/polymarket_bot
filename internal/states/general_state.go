package states

import (
	"crypto/rand"
	"math/big"

	actors_market "github.com/JaKoSTaWiND/polymarket_bot/internal/actors/market"
)

// GeneralState хранит общее состояние одного запущенного бота.
type GeneralState struct {
	BotID                     string                                    // Идентификатор экземпляра бота.
	BotStatus                 string                                    // Текущий статус жизненного цикла.
	Balance                   uint64                                    // Последний полученный баланс в миллионных долях pUSD. Для конкурентного чтения используйте atomic.LoadUint64. | 1pUSD -> 1_000_000
	OpenPositions             map[string]actors_market.PositionsByToken // Снимок позиций по ConditionID и токену, включая ожидающие погашения.
	MarketActorsByConditionID map[string]actors_market.MarketActor      // Акторы рынков по ConditionID.
	MaxTrackingMarkets        int                                       // Максимальное число разных рынков в реестре.
}

func GenerateBotID() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var id [16]byte
	limit := big.NewInt(int64(len(alphabet)))

	for i := range id {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		id[i] = alphabet[index.Int64()]
	}
	return string(id[:]), nil
}

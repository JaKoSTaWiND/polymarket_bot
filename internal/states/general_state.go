package states

import (
	"crypto/rand"
	"math/big"
)

type GeneralState struct {
	BotID     string
	BotStatus string
	Balance   int64 // Последний полученный баланс в миллионных долях pUSD. Для конкурентного чтения используйте atomic.LoadInt64.
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

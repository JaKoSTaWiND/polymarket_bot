package entities

type Token struct {
	AssetID string
	Outcome string
	Price   int64 // Цена в микро-pUSD за один токен.
}

package entities

import "time"

type Market struct {
	MarketID    string
	Slug        string
	ConditionID string
	Question    string
	EndDate     *time.Time // "2023-11-07T05:31:56Z"
	StartDate   *time.Time // "2023-11-07T05:31:56Z"
	Liquidity   int64      // Ликвидность рынка из Gamma в миллионных долях.
	Volume      int64      // Общий объём торгов из Gamma в миллионных долях.

	PrimaryToken   Token
	SecondaryToken Token
}

package entities

import "time"

type Market struct {
	MarketID    string
	Slug        string
	ConditionID string
	Question    string
	EndDate     *time.Time // "2023-11-07T05:31:56Z"
	StartDate   *time.Time // "2023-11-07T05:31:56Z"
	Liquidity   uint64
	Volume      uint64

	PrimaryToken   Token
	SecondaryToken Token
}

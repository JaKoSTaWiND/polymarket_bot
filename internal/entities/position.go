package entities

import "github.com/shopspring/decimal"

type Position struct {
	PositionID         string // генерирует сам бот
	AssetID            string
	ConditionID        string
	CurrentSize        decimal.Decimal // Количество шеирсов
	AvgPrice           uint64          // Средняя цена в шеирсах
	EntryCostUsdc      uint64          // Цена покупки
	EntryFeesUsdc      uint64          // Налог
	TotalCostUsdc      uint64          // Итоговая цена покупки
	CurrentPrice       uint64          // Цена в шеирсах сейчас
	CurrentValue       uint64          // Сколько стоит позиция в usdc сейчас
	TotalSize          decimal.Decimal // Количество шеирсов сейчас
	RealizedPnl        decimal.Decimal // Накопленная прибыль или убыток по уже закрытой части позиции
	UnrealizedPnl      decimal.Decimal // Прибыль или убыток по токенам, которые ещё в позиции. Текущая стоимость остатка минус его себестоимость без комиссий.
	TotalPnl           decimal.Decimal // Сумма RealizedPnl и UnrealizedPnl
	PercentPnl         decimal.Decimal // Доходность оставшихся токенов. UnrealizedPnl / entry_cost_usdc × 100
	PercentRealizedPnl decimal.Decimal // Поле для совместимости. Сравнивает текущую стоимость остатка с базой total_size × avg_price, где total_size — весь когда-либо купленный объём
}

package stock

type SubSetData struct {
	StockCode        string `json:"stock_code"`
	Type             string `json:"type"`
	Quantity         int    `json:"quantity,string"`
	ExecutedQuantity int    `json:"executed_quantity,string"`
	Price            int    `json:"price,string"`
	ExecutionPrice   int    `json:"execution_price,string"`
}

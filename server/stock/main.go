package stock

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/scizorman/go-ndjson"
	"log"
	"math"
	"os"
)

type DataStockServer struct {
	pb.UnimplementedDataStockServer
}

func NewDataStockServer() *DataStockServer {
	return &DataStockServer{}
}

func (s *DataStockServer) GetOneSummary(_ context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	return loadStockData(in.StockSymbol)
}

func loadStockData(stockCode string) (*pb.Stock, error) {
	directory, err := os.ReadDir("subsetdata")
	if err != nil {
		log.Fatalln("error reading directory", err.Error())
	}

	var (
		prevPrice  int64
		openPrice  int64
		highPrice  int64
		lowPrice   int64
		closePrice int64
		value      int64
		volume     int64
	)

	for _, file := range directory {
		data, err := os.ReadFile("subsetdata/" + file.Name())
		if err != nil {
			log.Fatalln("error reading file", err.Error())
		}

		var subsetData []SubSetData
		if err = ndjson.Unmarshal(data, &subsetData); err != nil {
			log.Fatalln("error unmarshalling data", err.Error())
		}

		for _, subsetDatum := range subsetData {
			if subsetDatum.StockCode == stockCode {
				// for first data
				if subsetDatum.Type == "A" && subsetDatum.Quantity == 0 && subsetDatum.ExecutedQuantity == 0 {
					prevPrice = int64(subsetDatum.Price)
					if prevPrice == 0 {
						prevPrice = int64(subsetDatum.ExecutionPrice)
					}
					continue
				}

				if subsetDatum.Type == "E" || subsetDatum.Type == "P" {
					// for first trade
					if openPrice == 0 {
						openPrice = int64(subsetDatum.Price)
						if openPrice == 0 {
							openPrice = int64(subsetDatum.ExecutionPrice)
						}
						continue
					}

					// for close price
					if subsetDatum.Price != 0 {
						closePrice = int64(subsetDatum.Price)
					} else if subsetDatum.ExecutionPrice != 0 {
						closePrice = int64(subsetDatum.ExecutionPrice)
					}

					// for high price
					if highPrice < closePrice {
						highPrice = closePrice
					}

					// for low price
					if lowPrice == 0 || lowPrice > closePrice {
						lowPrice = closePrice
					}

					// for volume
					volume += int64(subsetDatum.Quantity) + int64(subsetDatum.ExecutedQuantity)

					// for total price
					value += closePrice * (int64(subsetDatum.Quantity) + int64(subsetDatum.ExecutedQuantity))
				}
			}
		}
	}

	var avgPrice float64
	if volume != 0 {
		avgPrice = float64(value) / float64(volume)
	}

	return &pb.Stock{
		Symbol:     stockCode,
		PrevPrice:  prevPrice,
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closePrice,
		AvgPrice:   int64(math.Round(avgPrice)),
		Volume:     volume,
		Value:      value,
	}, nil
}

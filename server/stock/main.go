package stock

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/scizorman/go-ndjson"
	"log"
	"math"
	"os"
	"strings"
)

type DataStockServer struct {
	pb.UnimplementedDataStockServer
	publisher    pubsub.Publisher
	redisManager *db.Redis
}

func NewDataStockServer(
	publisher pubsub.Publisher,
	redisManager *db.Redis,
) *DataStockServer {
	return &DataStockServer{
		publisher:    publisher,
		redisManager: redisManager,
	}
}

func (s *DataStockServer) GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	var stock *pb.Stock
	return stock, s.redisManager.Get(ctx, strings.ToUpper(in.StockSymbol), &stock)
}

func (s *DataStockServer) LoadInitialData(ctx context.Context) error {
	directory, err := os.ReadDir("subsetdata")
	if err != nil {
		log.Fatalln("error reading directory", err.Error())
	}

	firstFile := directory[0]
	firstData, err := os.ReadFile("subsetdata/" + firstFile.Name())
	if err != nil {
		log.Fatalln("error reading file", err.Error())
	}

	var firstSubSetData []SubSetData
	if err = ndjson.Unmarshal(firstData, &firstSubSetData); err != nil {
		log.Fatalln("error unmarshalling data", err.Error())
	}

	mapStockData := make(map[string]*pb.Stock)
	var keys []string
	for _, subsetDatum := range firstSubSetData {
		if _, ok := mapStockData[subsetDatum.StockCode]; !ok {
			price := int64(subsetDatum.Price)
			mapStockData[subsetDatum.StockCode] = &pb.Stock{
				Symbol:    subsetDatum.StockCode,
				PrevPrice: price,
			}
			keys = append(keys, subsetDatum.StockCode)
		}
	}

	latestFileName := directory[len(directory)-1].Name()
	latestScannedRow := 0
	directory = directory[1:]
	var cleanSubSetData []SubSetData
	for _, file := range directory {
		data, err := os.ReadFile("subsetdata/" + file.Name())
		if err != nil {
			log.Fatalln("error reading file", err.Error())
		}

		var tempSubsetData []SubSetData
		if err = ndjson.Unmarshal(data, &tempSubsetData); err != nil {
			log.Fatalln("error unmarshalling data", err.Error())
		}

		for _, subsetDatum := range tempSubsetData {
			if subsetDatum.Type == "E" || subsetDatum.Type == "P" {
				cleanSubSetData = append(cleanSubSetData, subsetDatum)
			}
		}

		if file.Name() == latestFileName {
			latestScannedRow = len(tempSubsetData)
		}
	}

	for _, subsetDatum := range cleanSubSetData {
		closePrice := int64(subsetDatum.Price + subsetDatum.ExecutionPrice)
		quantity := int64(subsetDatum.Quantity + subsetDatum.ExecutedQuantity)
		mapStockData[subsetDatum.StockCode].ClosePrice = closePrice
		mapStockData[subsetDatum.StockCode].Volume += quantity
		mapStockData[subsetDatum.StockCode].Value += closePrice * quantity
		switch {
		case mapStockData[subsetDatum.StockCode].OpenPrice == 0:
			mapStockData[subsetDatum.StockCode].OpenPrice = int64(subsetDatum.Price) + int64(subsetDatum.ExecutionPrice)
			fallthrough
		case mapStockData[subsetDatum.StockCode].HighPrice < closePrice:
			mapStockData[subsetDatum.StockCode].HighPrice = closePrice
			fallthrough
		case mapStockData[subsetDatum.StockCode].LowPrice > closePrice:
			mapStockData[subsetDatum.StockCode].LowPrice = closePrice
		}
	}

	for _, key := range keys {
		if mapStockData[key].Volume != 0 {
			mapStockData[key].AvgPrice = mapStockData[key].Value / mapStockData[key].Volume
		}
		if err = s.redisManager.Set(ctx, key, mapStockData[key]); err != nil {
			log.Fatalln("error setting data", err.Error())
		}
	}

	if err = s.redisManager.Set(ctx, "latestScannedRow", latestScannedRow); err != nil {
		log.Fatalln("error setting data", err.Error())
	}

	if err = s.redisManager.Set(ctx, "latestFileName", latestFileName); err != nil {
		log.Fatalln("error setting data", err.Error())
	}

	return nil
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

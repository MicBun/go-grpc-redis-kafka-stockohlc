package main

import (
	"context"
	"fmt"
	pb "github.com/MicBun/go-grpc-redis-kafka-stockohlc/stock"
	"github.com/scizorman/go-ndjson"
	"google.golang.org/grpc"
	"log"
	"math"
	"net"
	"os"
	"sync"
)

type dataStockServer struct {
	pb.UnimplementedDataStockServer
	mu     sync.Mutex
	stocks []*pb.Stock
}

func (s *dataStockServer) GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	panic("implement me")
}

type SubSetData struct {
	StockCode        string `json:"stock_code"`
	Type             string `json:"type"`
	Quantity         int    `json:"quantity,string"`
	ExecutedQuantity int    `json:"executed_quantity,string"`
	Price            int    `json:"price,string"`
	ExecutionPrice   int    `json:"execution_price,string"`
}

func (s *dataStockServer) loadStockData(stockCode string) (pb.Stock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	return pb.Stock{
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

func newServer() *dataStockServer {
	s := &dataStockServer{}
	bbca, _ := s.loadStockData("TLKM")
	fmt.Println("bbca.Symbol", bbca.Symbol)
	fmt.Println("bbca.PrevPrice", bbca.PrevPrice)
	fmt.Println("bbca.OpenPrice", bbca.OpenPrice)
	fmt.Println("bbca.HighPrice", bbca.HighPrice)
	fmt.Println("bbca.LowPrice", bbca.LowPrice)
	fmt.Println("bbca.ClosePrice", bbca.ClosePrice)
	fmt.Println("bbca.Volume", bbca.Volume)
	fmt.Println("bbca.Value", bbca.Value)
	fmt.Println("bbca.AvgPrice", bbca.AvgPrice)
	return s
}

func main() {
	listen, err := net.Listen("tcp", ":1200")
	if err != nil {
		log.Fatalln("failed to listen", err.Error())
	}

	gRPCServer := grpc.NewServer()
	pb.RegisterDataStockServer(gRPCServer, newServer())

	log.Println("starting gRPC server at", listen.Addr().String())
	if err = gRPCServer.Serve(listen); err != nil {
		log.Fatalln("failed to serve", err.Error())
	}
}

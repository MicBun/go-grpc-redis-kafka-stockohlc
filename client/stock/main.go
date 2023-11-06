package stock

import (
	"context"
	"time"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock/pb"
)

type Stock struct{}

type Manager interface {
	GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error)
}

func NewManager() *Stock {
	return &Stock{}
}

func (*Stock) GetOneStockSummary(client pb.DataStockClient, stockSymbol string) (*pb.Stock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetOneSummary(ctx, &pb.GetOneSummaryRequest{StockSymbol: stockSymbol})
}

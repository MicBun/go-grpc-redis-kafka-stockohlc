package stock

import (
	"context"
	"strings"
	"sync"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/filesystem"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/pkg/errors"
	"github.com/scizorman/go-ndjson"
)

type DataStockManager interface {
	GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error)
	LoadInitialData(ctx context.Context) error
	UpdateStockOnFileUpdate(ctx context.Context, updatedFileName string) error
	UpdateStockOnFileCreate(ctx context.Context, createdFileName string) error
	SetMapStockData(ctx context.Context, mapStockData map[string]*pb.Stock) error
}

type DataStockServer struct {
	pb.UnimplementedDataStockServer
	publisher    pubsub.Publisher
	redisManager db.RedisManager
	mu           *sync.Mutex
	fileSystem   filesystem.FS
}

func NewDataStockServer(
	publisher pubsub.Publisher,
	redisManager db.RedisManager,
	mu *sync.Mutex,
	fileSystem filesystem.FS,
) *DataStockServer {
	return &DataStockServer{
		publisher:    publisher,
		redisManager: redisManager,
		mu:           mu,
		fileSystem:   fileSystem,
	}
}

func (s *DataStockServer) GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	var stock *pb.Stock
	return stock, s.redisManager.Get(ctx, strings.ToUpper(in.StockSymbol), &stock)
}

func (s *DataStockServer) LoadInitialData(ctx context.Context) error {
	directory, err := s.fileSystem.ReadDir("subsetdata")
	if err != nil {
		return errors.WithStack(err)
	}

	firstFile := directory[0]
	firstData, err := s.fileSystem.ReadFile("subsetdata/" + firstFile.Name())
	if err != nil {
		return errors.WithStack(err)
	}

	var firstSubSetData []SubSetData
	if err = ndjson.Unmarshal(firstData, &firstSubSetData); err != nil {
		return errors.WithStack(err)
	}

	mapStockData := make(map[string]*pb.Stock)
	for _, subsetDatum := range firstSubSetData {
		if _, ok := mapStockData[subsetDatum.StockCode]; !ok {
			mapStockData[subsetDatum.StockCode] = &pb.Stock{
				Symbol:    subsetDatum.StockCode,
				PrevPrice: int64(subsetDatum.Price),
			}
		}
	}

	latestFileName := directory[len(directory)-1].Name()
	latestScannedRow := 0
	directory = directory[1:]
	var cleanSubSetData []SubSetData
	for _, file := range directory {
		data, errRead := s.fileSystem.ReadFile("subsetdata/" + file.Name())
		if errRead != nil {
			return errors.WithStack(errRead)
		}

		var dirtySubsetData []SubSetData
		if err = ndjson.Unmarshal(data, &dirtySubsetData); err != nil {
			return errors.WithStack(err)
		}

		cleanSubSetData = append(cleanSubSetData, getCleansedSubSetData(dirtySubsetData)...)
		if file.Name() == latestFileName {
			latestScannedRow = len(dirtySubsetData)
		}
	}

	if err = s.SetMapStockData(ctx, calculateOHLC(cleanSubSetData, mapStockData)); err != nil {
		return errors.WithStack(err)
	}
	return setLatestScannedRowAndFileName(ctx, s.redisManager, latestScannedRow, latestFileName)
}

func (s *DataStockServer) UpdateStockOnFileUpdate(ctx context.Context, updatedFileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latestScannedRow int
	if err := s.redisManager.Get(ctx, db.LatestScannedRow, &latestScannedRow); err != nil {
		return errors.WithStack(err)
	}

	data, err := s.fileSystem.ReadFile(updatedFileName)
	if err != nil {
		return errors.WithStack(err)
	}

	var dirtySubsetData []SubSetData
	if err = ndjson.Unmarshal(data, &dirtySubsetData); err != nil {
		return errors.WithStack(err)
	}

	if len(dirtySubsetData) == latestScannedRow {
		return nil
	}
	dirtySubsetData = dirtySubsetData[latestScannedRow:]

	var keys []string
	var mapStockData = make(map[string]*pb.Stock)
	cleanSubSetData := getCleansedSubSetData(dirtySubsetData)

	for _, subsetDatum := range cleanSubSetData {
		if _, ok := mapStockData[subsetDatum.StockCode]; !ok {
			mapStockData[subsetDatum.StockCode] = &pb.Stock{
				Symbol: subsetDatum.StockCode,
			}
			keys = append(keys, subsetDatum.StockCode)
		}
	}

	for _, key := range keys {
		if err = s.redisManager.Get(ctx, key, mapStockData[key]); err != nil {
			return errors.WithStack(err)
		}
	}

	if err = s.SetMapStockData(ctx, calculateOHLC(cleanSubSetData, mapStockData)); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(s.redisManager.Set(ctx, db.LatestScannedRow, latestScannedRow+len(dirtySubsetData)))
}

func (s *DataStockServer) UpdateStockOnFileCreate(ctx context.Context, createdFileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latestFileName string
	if err := s.redisManager.Get(ctx, db.LatestFileName, &latestFileName); err != nil {
		return errors.WithStack(err)
	}
	if createdFileName == ("subsetdata/" + latestFileName) {
		return nil
	}
	directory, err := s.fileSystem.ReadDir("subsetdata")
	if err != nil {
		return errors.WithStack(err)
	}
	for i := len(directory) - 1; i >= 0; i-- {
		if ("subsetdata/" + directory[i].Name()) == createdFileName {
			directory = directory[i:]
			break
		}
	}

	var (
		cleanSubSetData  []SubSetData
		latestScannedRow int
	)
	for _, file := range directory {
		data, errRead := s.fileSystem.ReadFile("subsetdata/" + file.Name())
		if errRead != nil {
			return errors.WithStack(errRead)
		}
		var dirtySubsetData []SubSetData
		if err = ndjson.Unmarshal(data, &dirtySubsetData); err != nil {
			return errors.WithStack(err)
		}
		cleanSubSetData = append(cleanSubSetData, getCleansedSubSetData(dirtySubsetData)...)
		if directory[len(directory)-1].Name() == file.Name() {
			latestScannedRow = len(dirtySubsetData)
		}
	}

	var mapStockData = make(map[string]*pb.Stock)
	var keys []string
	for _, subsetDatum := range cleanSubSetData {
		if _, ok := mapStockData[subsetDatum.StockCode]; !ok {
			mapStockData[subsetDatum.StockCode] = &pb.Stock{
				Symbol: subsetDatum.StockCode,
			}
			keys = append(keys, subsetDatum.StockCode)
		}
	}
	for _, key := range keys {
		if err = s.redisManager.Get(ctx, key, mapStockData[key]); err != nil {
			return errors.WithStack(err)
		}
	}
	mapStockData = calculateOHLC(cleanSubSetData, mapStockData)
	if err = s.SetMapStockData(ctx, mapStockData); err != nil {
		return errors.WithStack(err)
	}
	return setLatestScannedRowAndFileName(ctx, s.redisManager, latestScannedRow, latestFileName)
}

func (s *DataStockServer) SetMapStockData(ctx context.Context, mapStockData map[string]*pb.Stock) error {
	for key, value := range mapStockData {
		if value.Volume != 0 {
			value.AvgPrice = value.Value / value.Volume
		}

		if err := s.redisManager.Set(ctx, key, value); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func calculateOHLC(cleanSubSetData []SubSetData, mapStockData map[string]*pb.Stock) map[string]*pb.Stock {
	for _, subsetDatum := range cleanSubSetData {
		closePrice := int64(subsetDatum.Price + subsetDatum.ExecutionPrice)
		quantity := int64(subsetDatum.Quantity + subsetDatum.ExecutedQuantity)
		if quantity == 0 {
			mapStockData[subsetDatum.StockCode] = new(pb.Stock)
			mapStockData[subsetDatum.StockCode].PrevPrice = closePrice
			continue
		}
		mapStockData[subsetDatum.StockCode].ClosePrice = closePrice
		mapStockData[subsetDatum.StockCode].Volume += quantity
		mapStockData[subsetDatum.StockCode].Value += closePrice * quantity
		switch {
		case mapStockData[subsetDatum.StockCode].OpenPrice == 0:
			mapStockData[subsetDatum.StockCode].OpenPrice = int64(subsetDatum.Price) + int64(subsetDatum.ExecutionPrice)
			mapStockData[subsetDatum.StockCode].HighPrice = int64(subsetDatum.Price) + int64(subsetDatum.ExecutionPrice)
			mapStockData[subsetDatum.StockCode].LowPrice = int64(subsetDatum.Price) + int64(subsetDatum.ExecutionPrice)
		case mapStockData[subsetDatum.StockCode].HighPrice < closePrice:
			mapStockData[subsetDatum.StockCode].HighPrice = closePrice
		case mapStockData[subsetDatum.StockCode].LowPrice > closePrice:
			mapStockData[subsetDatum.StockCode].LowPrice = closePrice
		}
	}

	return mapStockData
}

func setLatestScannedRowAndFileName(ctx context.Context, redisManager db.RedisManager, latestScannedRow int, latestFileName string) error {
	if err := redisManager.Set(ctx, db.LatestScannedRow, latestScannedRow); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(redisManager.Set(ctx, db.LatestFileName, latestFileName))
}

func getCleansedSubSetData(dirtySubsetData []SubSetData) []SubSetData {
	var cleanSubSetData []SubSetData
	for _, subsetDatum := range dirtySubsetData {
		if subsetDatum.Type == "E" || subsetDatum.Type == "P" || (subsetDatum.Quantity == 0 && subsetDatum.ExecutedQuantity == 0) {
			cleanSubSetData = append(cleanSubSetData, subsetDatum)
		}
	}
	return cleanSubSetData
}

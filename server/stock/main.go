package stock

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/scizorman/go-ndjson"
	"log"
	"os"
	"strings"
	"sync"
)

type DataStockServer struct {
	pb.UnimplementedDataStockServer
	publisher    pubsub.Publisher
	redisManager *db.Redis
	mu           *sync.Mutex
}

func NewDataStockServer(
	publisher pubsub.Publisher,
	redisManager *db.Redis,
	mu *sync.Mutex,
) *DataStockServer {
	return &DataStockServer{
		publisher:    publisher,
		redisManager: redisManager,
		mu:           mu,
	}
}

func (s *DataStockServer) GetOneSummary(ctx context.Context, in *pb.GetOneSummaryRequest) (*pb.Stock, error) {
	var stock *pb.Stock
	return stock, s.redisManager.Get(ctx, strings.ToUpper(in.StockSymbol), &stock)
}

func (s *DataStockServer) LoadInitialData(ctx context.Context) error {
	directory, err := os.ReadDir("subsetdata")
	if err != nil {
		return errors.WithStack(err)
	}

	firstFile := directory[0]
	firstData, err := os.ReadFile("subsetdata/" + firstFile.Name())
	if err != nil {
		return errors.WithStack(err)
	}

	var firstSubSetData []SubSetData
	if err = ndjson.Unmarshal(firstData, &firstSubSetData); err != nil {
		return errors.WithStack(err)
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
		data, errRead := os.ReadFile("subsetdata/" + file.Name())
		if err != nil {
			return errors.WithStack(errRead)
		}

		var dirtySubsetData []SubSetData
		if err = ndjson.Unmarshal(data, &dirtySubsetData); err != nil {
			return errors.WithStack(err)
		}

		for _, subsetDatum := range dirtySubsetData {
			if subsetDatum.Type == "E" || subsetDatum.Type == "P" {
				cleanSubSetData = append(cleanSubSetData, subsetDatum)
			}
		}

		if file.Name() == latestFileName {
			latestScannedRow = len(dirtySubsetData)
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
			return errors.WithStack(err)
		}
	}

	if err = s.redisManager.Set(ctx, db.LatestScannedRow, latestScannedRow); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(s.redisManager.Set(ctx, db.LatestFileName, latestFileName))
}

func (s *DataStockServer) StartDirectoryMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Close()

	if err = watcher.Add("subsetdata"); err != nil {
		log.Fatalf("Error adding watcher: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if err = s.publisher.Publish(ctx, pubsub.TopicFileCreated, event.Name); err != nil {
					log.Fatalln("error publishing message", err.Error())
				}
			} else if event.Has(fsnotify.Write) {
				if err = s.publisher.Publish(ctx, pubsub.TopicFileUpdated, event.Name); err != nil {
					log.Fatalln("error publishing message", err.Error())
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error: %v", err)
		}
	}
}

func (s *DataStockServer) UpdateStockOnFileUpdate(ctx context.Context, updatedFileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latestScannedRow int
	if err := s.redisManager.Get(ctx, "latestScannedRow", &latestScannedRow); err != nil {
		return errors.WithStack(err)
	}

	data, err := os.ReadFile(updatedFileName)
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

	var cleanSubSetData []SubSetData
	var keys []string
	var mapStockData = make(map[string]*pb.Stock)
	for _, subsetDatum := range dirtySubsetData {
		if subsetDatum.Type == "E" || subsetDatum.Type == "P" {
			cleanSubSetData = append(cleanSubSetData, subsetDatum)
			if _, ok := mapStockData[subsetDatum.StockCode]; !ok {
				mapStockData[subsetDatum.StockCode] = &pb.Stock{
					Symbol: subsetDatum.StockCode,
				}
				keys = append(keys, subsetDatum.StockCode)
			}

			latestScannedRow--
		}
	}

	for _, key := range keys {
		if err = s.redisManager.Get(ctx, key, mapStockData[key]); err != nil {
			return errors.WithStack(err)
		}
	}

	for _, subsetDatum := range cleanSubSetData[latestScannedRow+1:] {
		closePrice := int64(subsetDatum.Price + subsetDatum.ExecutionPrice)
		quantity := int64(subsetDatum.Quantity + subsetDatum.ExecutedQuantity)
		mapStockData[subsetDatum.StockCode].ClosePrice = closePrice
		mapStockData[subsetDatum.StockCode].Volume += quantity
		mapStockData[subsetDatum.StockCode].Value += closePrice * quantity
		switch {
		case mapStockData[subsetDatum.StockCode].OpenPrice == 0:
			mapStockData[subsetDatum.StockCode].OpenPrice = closePrice
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
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(s.redisManager.Set(ctx, "latestScannedRow", len(dirtySubsetData)-1))
}

func (s *DataStockServer) UpdateStockOnFileCreate(ctx context.Context, createdFileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latestFileName string
	if err := s.redisManager.Get(ctx, "latestFileName", &latestFileName); err != nil {
		return errors.WithStack(err)
	}

	if createdFileName == ("subsetdata/" + latestFileName) {
		return nil
	}

	directory, err := os.ReadDir("subsetdata")
	if err != nil {
		log.Fatalln("error reading directory", err.Error())
	}

	for i := len(directory) - 1; i >= 0; i-- {
		if ("subsetdata/" + directory[i].Name()) == createdFileName {
			directory = directory[i:]
			break
		}
	}

	var cleanSubSetData []SubSetData
	var latestScannedRow int
	for _, file := range directory {
		data, errRead := os.ReadFile("subsetdata/" + file.Name())
		if err != nil {
			return errors.WithStack(errRead)
		}

		var dirtySubsetData []SubSetData
		if err = ndjson.Unmarshal(data, &dirtySubsetData); err != nil {
			return errors.WithStack(err)
		}

		for _, subsetDatum := range dirtySubsetData {
			if subsetDatum.Type == "E" || subsetDatum.Type == "P" {
				cleanSubSetData = append(cleanSubSetData, subsetDatum)
			}
		}

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
			return errors.WithStack(err)
		}
	}

	if err = s.redisManager.Set(ctx, db.LatestScannedRow, latestScannedRow); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(s.redisManager.Set(ctx, db.LatestFileName, latestFileName))
}

package stock

import (
	"context"
	"io/fs"
	"sync"
	"testing"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/filesystem"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testStockInstance struct {
	stock      *DataStockServer
	publisher  *pubsub.MockPublisher
	subscriber *pubsub.MockSubscriber
	redis      *db.MockRedisManager
	fileSystem *filesystem.MockFS
	mu         *sync.Mutex
}

func newTestStockInstance() *testStockInstance {
	publisher := new(pubsub.MockPublisher)
	subscriber := new(pubsub.MockSubscriber)
	redis := new(db.MockRedisManager)
	fileSystem := new(filesystem.MockFS)
	mu := new(sync.Mutex)

	return &testStockInstance{
		stock:      NewDataStockServer(publisher, redis, mu, fileSystem),
		publisher:  publisher,
		subscriber: subscriber,
		redis:      redis,
		fileSystem: fileSystem,
		mu:         mu,
	}
}

func TestDataStockServer_GetOneSummary(t *testing.T) {
	t.Run("Success - it should return stock", func(t *testing.T) {
		instance := newTestStockInstance()
		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		_, err := instance.stock.GetOneSummary(context.Background(), &pb.GetOneSummaryRequest{})
		assert.NoError(t, err)
	})
}

func TestDataStockServer_LoadInitialData(t *testing.T) {
	t.Run("Success - it should load initial data", func(t *testing.T) {
		instance := newTestStockInstance()

		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(
			`{"stock_code":"test","volume":"1","type":"E","price":"50"}
			{"stock_code":"test","volume":"1","type":"E","price":"55","quantity":"1"}
			{"stock_code":"test","volume":"1","type":"E","price":"45","quantity":"1"}
			{"stock_code":"test","volume":"1","type":"E","price":"60","quantity":"1"}`,
		), nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, instance.stock.LoadInitialData(context.Background()))
	})

	t.Run("Error - it should return error when file system read dir failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return(nil, assert.AnError)

		assert.EqualError(t, instance.stock.LoadInitialData(context.Background()), assert.AnError.Error())
	})

	t.Run("Error - it should return error when file system read file failed", func(t *testing.T) {
		instance := newTestStockInstance()

		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return(nil, assert.AnError)

		assert.EqualError(t, instance.stock.LoadInitialData(context.Background()), assert.AnError.Error())
	})

	t.Run("Error - it should error if given wrong json format", func(t *testing.T) {
		instance := newTestStockInstance()

		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(`{"symbol":"test","volume":1`), nil)

		assert.EqualError(t, instance.stock.LoadInitialData(context.Background()), "unexpected end of JSON input")
	})

	t.Run("Error - it should return error when redis set failed", func(t *testing.T) {
		instance := newTestStockInstance()

		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(`{"stock_code":"test","volume":1}`), nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.LoadInitialData(context.Background()), assert.AnError.Error())
	})
}

func TestDataStockServer_UpdateStockOnFileCreate(t *testing.T) {
	t.Run("Success - it should update stock on file create", func(t *testing.T) {
		instance := newTestStockInstance()
		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(
			`{"symbol":"test","volume":1,"type":"E"}`),
			nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
			&mockDirEntry,
		}, nil)

		assert.NoError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"subsetdata/test"`))
	})

	t.Run("Error - it should return error when redis get failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"test"`), assert.AnError.Error())
	})

	t.Run("Error - it should return error when file system read dir failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return(nil, assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"test"`), assert.AnError.Error())
	})

	t.Run("Error - it should return error when file system read file failed", func(t *testing.T) {
		instance := newTestStockInstance()
		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte{}, assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"test"`), assert.AnError.Error())
	})

	t.Run("Error - it should return error when json unmarshal failed", func(t *testing.T) {
		instance := newTestStockInstance()
		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(`{"symbol":"test","volume":1`), nil)

		assert.EqualError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"test"`), "unexpected end of JSON input")
	})

	t.Run("Error - it should return error when redis set failed", func(t *testing.T) {
		instance := newTestStockInstance()
		mockDirEntry := filesystem.MockDirEntry{}
		mockDirEntry.On("Name").Return("test")

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{
			&mockDirEntry,
		}, nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(
			`{"symbol":"test","volume":1,"type":"E"}`),
			nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileCreate(context.Background(), `"test"`), assert.AnError.Error())
	})
}

func TestDataStockServer_SetMapStockData(t *testing.T) {
	t.Run("Success - it should set map stock data", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		assert.NoError(t, instance.stock.SetMapStockData(context.Background(), map[string]*pb.Stock{
			"test": {
				Symbol: "test",
				Volume: 1,
			},
		}))
	})

	t.Run("Error - it should return error when redis set failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.SetMapStockData(context.Background(), map[string]*pb.Stock{
			"test": {
				Symbol: "test",
				Volume: 1,
			},
		}), assert.AnError.Error())
	})
}

func TestDataStockServer_UpdateStockOnFileUpdate(t *testing.T) {
	t.Run("Success - it should update stock on file update", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(
			`{"symbol":"test","volume":1,"type":"E"}`),
			nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadDir(mock.Anything).Return([]fs.DirEntry{}, nil)

		assert.NoError(t, instance.stock.UpdateStockOnFileUpdate(context.Background(), `"test"`))
	})

	t.Run("Error - it should return error when redis get failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileUpdate(context.Background(), `"test"`), assert.AnError.Error())
	})

	t.Run("Error - it should return error when file system read file failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return(nil, assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileUpdate(context.Background(), `"test"`), assert.AnError.Error())
	})

	t.Run("Error - it should return error when json unmarshal failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(`{"symbol":"test","volume":1`), nil)

		assert.EqualError(t, instance.stock.UpdateStockOnFileUpdate(context.Background(), `"test"`), "unexpected end of JSON input")
	})

	t.Run("Error - it should return error when redis set failed", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.redis.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		instance.fileSystem.EXPECT().ReadFile(mock.Anything).Return([]byte(
			`{"symbol":"test","volume":1,"type":"E"}`),
			nil)
		instance.redis.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)

		assert.EqualError(t, instance.stock.UpdateStockOnFileUpdate(context.Background(), `"test"`), assert.AnError.Error())
	})
}

package stock

import (
	"testing"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testStockInstance struct {
	stock        *Stock
	pbMockClient *pb.MockDataStockClient
}

func newTestStockInstance() *testStockInstance {
	pbMockClient := new(pb.MockDataStockClient)
	return &testStockInstance{
		stock:        NewManager(),
		pbMockClient: pbMockClient,
	}
}

func Test_getOneStockSummary(t *testing.T) {
	t.Run("Success - it should return stock", func(t *testing.T) {
		instance := newTestStockInstance()

		instance.pbMockClient.EXPECT().GetOneSummary(mock.Anything, mock.Anything).Return(&pb.Stock{}, nil)
		_, err := instance.stock.GetOneStockSummary(instance.pbMockClient, "test")
		assert.NoError(t, err)
	})
}

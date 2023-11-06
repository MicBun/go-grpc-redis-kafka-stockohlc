package message

import (
	"context"
	"testing"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testMessageInstance struct {
	stockManager *stock.MockDataStockManager
	message      *Message
}

func newTestMessageInstance() *testMessageInstance {
	stockManager := new(stock.MockDataStockManager)
	message := NewMessage(stockManager)
	return &testMessageInstance{
		stockManager: stockManager,
		message:      message,
	}
}

func TestMessage_FileCreated(t *testing.T) {
	t.Run("Success - it should update stock on file create", func(t *testing.T) {
		instance := newTestMessageInstance()

		instance.stockManager.EXPECT().UpdateStockOnFileCreate(mock.Anything, mock.Anything).Return(nil)
		err := instance.message.FileCreated(context.Background(), &message.Message{})
		assert.NoError(t, err)
	})
}

func TestMessage_FileUpdated(t *testing.T) {
	t.Run("Success - it should update stock on file update", func(t *testing.T) {
		instance := newTestMessageInstance()

		instance.stockManager.EXPECT().UpdateStockOnFileUpdate(mock.Anything, mock.Anything).Return(nil)
		err := instance.message.FileUpdated(context.Background(), &message.Message{})
		assert.NoError(t, err)
	})
}

func TestMessage_Load(t *testing.T) {
	t.Run("Success - it should subscribe to topics", func(t *testing.T) {
		instance := newTestMessageInstance()

		subscriber := new(pubsub.MockSubscriber)
		subscriber.EXPECT().Subscribe(pubsub.TopicFileCreated, mock.Anything)
		subscriber.EXPECT().Subscribe(pubsub.TopicFileUpdated, mock.Anything)

		instance.message.Load(subscriber)
	})
}

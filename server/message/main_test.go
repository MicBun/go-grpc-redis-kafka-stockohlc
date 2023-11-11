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

func TestMessage_FileUpdated(t *testing.T) {
	t.Run("Success - it should update stock on file update", func(t *testing.T) {
		instance := newTestMessageInstance()

		instance.stockManager.EXPECT().UpdateStockOnFileUpdate(mock.Anything, mock.Anything).Return(nil)
		err := instance.message.FileUpdated(context.Background(), &message.Message{
			Payload: []byte(`{"file_name":"test", "is_create":false}`),
		})
		assert.NoError(t, err)
	})

	t.Run("Success - it should update stock on file create", func(t *testing.T) {
		instance := newTestMessageInstance()

		instance.stockManager.EXPECT().UpdateStockOnFileCreate(mock.Anything, mock.Anything).Return(nil)
		err := instance.message.FileUpdated(context.Background(), &message.Message{
			Payload: []byte(`{"file_name":"test", "is_create":true}`),
		})
		assert.NoError(t, err)
	})

	t.Run("Error - it should return error on unmarshal", func(t *testing.T) {
		instance := newTestMessageInstance()

		err := instance.message.FileUpdated(context.Background(), &message.Message{
			Payload: []byte(`{"file_name":"test", "is_create":`),
		})
		assert.EqualError(t, err, "unexpected end of JSON input")
	})
}

func TestMessage_Load(t *testing.T) {
	t.Run("Success - it should subscribe to topics", func(t *testing.T) {
		instance := newTestMessageInstance()

		subscriber := new(pubsub.MockSubscriber)
		subscriber.EXPECT().Subscribe(pubsub.TopicFileUpdated, mock.Anything)

		instance.message.Load(subscriber)
	})
}

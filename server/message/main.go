package message

import (
	"context"
	"encoding/json"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Message struct {
	stockManager stock.DataStockManager
}

func NewMessage(
	stockManager stock.DataStockManager,
) *Message {
	return &Message{
		stockManager: stockManager,
	}
}

func (h *Message) FileUpdated(ctx context.Context, msg *message.Message) error {
	defer msg.Ack()

	body := new(FileUpdatedSchemaMessage)
	if err := json.Unmarshal(msg.Payload, body); err != nil {
		return err
	}

	if body.IsCreate {
		return h.stockManager.UpdateStockOnFileCreate(ctx, body.FileName)
	} else {
		return h.stockManager.UpdateStockOnFileUpdate(ctx, body.FileName)
	}
}

func (h *Message) Load(subscriber pubsub.Subscriber) {
	subscriber.Subscribe(pubsub.TopicFileUpdated, h.FileUpdated)
}

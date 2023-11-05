package message

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Message struct {
	stockManager *stock.DataStockServer
}

func NewMessage(
	stockManager *stock.DataStockServer,
) *Message {
	return &Message{
		stockManager: stockManager,
	}
}

func (h *Message) FileCreated(ctx context.Context, msg *message.Message) error {
	defer msg.Ack()
	return h.stockManager.UpdateStockOnFileCreate(ctx, string(msg.Payload)[1:len(msg.Payload)-1])
}

func (h *Message) FileUpdated(ctx context.Context, msg *message.Message) error {
	defer msg.Ack()
	return h.stockManager.UpdateStockOnFileUpdate(ctx, string(msg.Payload)[1:len(msg.Payload)-1])
}

func (h *Message) Load(subscriber pubsub.Subscriber) {
	subscriber.Subscribe(pubsub.TopicFileCreated, h.FileCreated)
	subscriber.Subscribe(pubsub.TopicFileUpdated, h.FileUpdated)
}

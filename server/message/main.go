package message

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/ThreeDotsLabs/watermill/message"
	"log"
)

type Message struct {
}

func NewMessage() *Message {
	return &Message{}
}

func (h *Message) SayHello(ctx context.Context, msg *message.Message) error {
	defer msg.Ack()
	log.Println("message received", msg)
	return nil
}

func (h *Message) Load(subscriber pubsub.Subscriber) {
	subscriber.Subscribe(pubsub.TopicExample, h.SayHello)
}

package pubsub

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	TopicExample = "grpcserver_example"
)

type SubscriberHandler func(context.Context, *message.Message) error

type ExceptionHandler interface {
	Handle(context context.Context, topic string, message *message.Message, err error)
}

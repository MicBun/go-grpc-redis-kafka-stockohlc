package pubsub

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	TopicFileUpdated = "grpcserver_file_updated"
)

type SubscriberHandler func(context.Context, *message.Message) error

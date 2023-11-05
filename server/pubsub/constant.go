package pubsub

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	TopicFileUpdated = "grpcserver_file_updated"
	TopicFileCreated = "grpcserver_file_created"
)

type SubscriberHandler func(context.Context, *message.Message) error

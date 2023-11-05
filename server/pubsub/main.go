package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

type Subscriber interface {
	Subscribe(topic string, handler SubscriberHandler)
	Start(ctx context.Context) error
	Close() error
}

type Publisher interface {
	Publish(ctx context.Context, topic string, message any) error
	Close() error
}

var watermillPublisherClient *kafka.Publisher

type WatermillPublisher struct {
	client *kafka.Publisher
}

func NewWatermillPublisher() (*WatermillPublisher, error) {
	if watermillPublisherClient == nil {
		host := os.Getenv("KAFKA_HOST")
		port := os.Getenv("KAFKA_PORT")
		url := host + ":" + fmt.Sprint(port)

		client, err := kafka.NewPublisher(
			kafka.PublisherConfig{
				Brokers:   []string{url},
				Marshaler: kafka.DefaultMarshaler{},
			},
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			return &WatermillPublisher{}, err
		}

		watermillPublisherClient = client
	}

	return &WatermillPublisher{watermillPublisherClient}, nil
}

func (p *WatermillPublisher) Publish(ctx context.Context, topic string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.client.Publish(topic, message.NewMessage(watermill.NewUUID(), data))
}

func (p *WatermillPublisher) Close() error {
	return p.client.Close()
}

var watermillSubscriberClient *kafka.Subscriber

type WatermillSubscriber struct {
	client     *kafka.Subscriber
	handlerMap map[string][]SubscriberHandler
}

func NewWatermillSubscriber() (*WatermillSubscriber, error) {
	if watermillSubscriberClient == nil {
		host := os.Getenv("KAFKA_HOST")
		port := os.Getenv("KAFKA_PORT")
		url := host + ":" + fmt.Sprint(port)
		consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")

		client, err := kafka.NewSubscriber(
			kafka.SubscriberConfig{
				Brokers:       []string{url},
				Unmarshaler:   kafka.DefaultMarshaler{},
				ConsumerGroup: consumerGroup,
			},
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			return &WatermillSubscriber{}, errors.WithStack(err)
		}

		watermillSubscriberClient = client
	}

	handlerMap := make(map[string][]SubscriberHandler)
	return &WatermillSubscriber{watermillSubscriberClient, handlerMap}, nil
}

func (s *WatermillSubscriber) Subscribe(topic string, handler SubscriberHandler) {
	if _, ok := s.handlerMap[topic]; !ok {
		s.handlerMap[topic] = []SubscriberHandler{}
	}

	s.handlerMap[topic] = append(s.handlerMap[topic], handler)
}

func (s *WatermillSubscriber) Start(ctx context.Context) error {
	for topic, handlers := range s.handlerMap {
		messages, err := s.client.Subscribe(ctx, topic)
		if err != nil {
			return errors.WithStack(err)
		}

		go s.processMessages(topic, messages, handlers)
	}

	return nil
}

func (s *WatermillSubscriber) Close() error {
	return s.client.Close()
}

func (s *WatermillSubscriber) processMessages(
	topic string,
	messages <-chan *message.Message,
	handlers []SubscriberHandler,
) {
	for msg := range messages {
		for _, handler := range handlers {
			go s.executeHandler(topic, handler, msg)
		}
	}
}

func (s *WatermillSubscriber) executeHandler(
	topic string,
	handler SubscriberHandler,
	msg *message.Message,
) {
	var listenerExecutionTimeoutInS = 30 * time.Second
	ctx, cancelCtx := context.WithTimeout(context.Background(), listenerExecutionTimeoutInS)
	defer cancelCtx()
	defer s.recoverFromPanic(ctx, topic, msg)

	if err := handler(ctx, msg); err != nil {
		log.Println("error executing handler", errors.WithStack(err))
	}
}

func (s *WatermillSubscriber) recoverFromPanic(ctx context.Context, topic string, msg *message.Message) {
	if r := recover(); r != nil {
		log.Println("panic recovered", errors.WithStack(fmt.Errorf("%v", r)))
	}
}

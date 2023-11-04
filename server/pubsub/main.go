package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

type SubscriberHandler func(context.Context, *message.Message) error

type Publisher interface {
	Publish(topic string, message any) error
	Close() error
}

type Subscriber interface {
	Subscribe(topic string, handler SubscriberHandler)
	Start() error
	Close() error
}

type ExceptionHandler interface {
	Handle(ctx context.Context, topic string, message *message.Message, err error)
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
		fmt.Println("url", url)

		client, err := kafka.NewPublisher(
			kafka.PublisherConfig{
				Brokers:   []string{url},
				Marshaler: kafka.DefaultMarshaler{},
			},
			watermill.NewStdLogger(true, false),
		)

		if err != nil {
			return &WatermillPublisher{}, err
		}

		watermillPublisherClient = client
	}

	return &WatermillPublisher{watermillPublisherClient}, nil
}

func (p *WatermillPublisher) Publish(topic string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	newMessage := message.NewMessage(watermill.NewUUID(), data)
	return p.client.Publish(topic, newMessage)
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

func (s *WatermillSubscriber) Start() error {
	var listenerExecutionTimeoutInS = 30 * time.Second

	for topic, handlers := range s.handlerMap {
		ctx := context.Background()
		messages, err := s.client.Subscribe(ctx, topic)
		if err != nil {
			return errors.WithStack(err)
		}

		go func(topic string, handlers []SubscriberHandler) {
			for msg := range messages {
				for _, handler := range handlers {
					go func(ctx context.Context, msg *message.Message, handler SubscriberHandler) {
						ctx, cancelCtx := context.WithTimeout(ctx, listenerExecutionTimeoutInS)
						defer cancelCtx()
						defer s.recoverFromPanic()

						if err = handler(ctx, msg); err != nil {
							log.Fatalln("error handling message", err.Error())
						}
					}(ctx, msg, handler)
				}
			}
		}(topic, handlers)
	}

	return nil
}

func (s *WatermillSubscriber) Close() error {
	return s.client.Close()
}

func (s *WatermillSubscriber) recoverFromPanic() {
	if r := recover(); r != nil {
		log.Fatalln("panic recovered:", r)
	}
}

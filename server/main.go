package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/message"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	subscriber pubsub.Subscriber
	publisher  pubsub.Publisher
	message    *message.Message
	redisDB    *db.Redis
	mu         *sync.Mutex
}

type NewAppArgs struct {
	gRPCServer *grpc.Server
	subscriber pubsub.Subscriber
	publisher  pubsub.Publisher
	message    *message.Message
	redisDB    *db.Redis
	mu         *sync.Mutex
}

func NewApp(
	args NewAppArgs,
) *App {
	app := &App{
		gRPCServer: args.gRPCServer,
		subscriber: args.subscriber,
		publisher:  args.publisher,
		message:    args.message,
		redisDB:    args.redisDB,
		mu:         args.mu,
	}
	pb.RegisterDataStockServer(app.gRPCServer, stock.NewDataStockServer(app.publisher, app.redisDB, app.mu))
	app.message.Load(app.subscriber)

	return app
}

func (a *App) Serve(ctx context.Context) error {
	if err := a.subscriber.Start(ctx); err != nil {
		return errors.Wrap(err, "error starting subscriber")
	}

	listen, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		return errors.WithStack(err)
	}

	return a.gRPCServer.Serve(listen)
}

func (a *App) Close() error {
	if err := a.subscriber.Close(); err != nil {
		return errors.Wrap(err, "error closing subscriber")
	}

	if err := a.publisher.Close(); err != nil {
		return errors.Wrap(err, "error closing publisher")
	}

	a.gRPCServer.Stop()
	return nil
}

func initApp() *App {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalln("Server failed to load env file", err.Error())
	}

	subscriber, err := pubsub.NewWatermillSubscriber()
	if err != nil {
		log.Fatalln("error creating watermill subscriber", err.Error())
	}

	publisher, err := pubsub.NewWatermillPublisher()
	if err != nil {
		log.Fatalln("error creating watermill publisher", err.Error())
	}

	mu := new(sync.Mutex)
	redisDB := db.NewRedisManager(db.NewRedis())
	stockManager := stock.NewDataStockServer(publisher, redisDB, mu)

	return NewApp(
		NewAppArgs{
			gRPCServer: grpc.NewServer(),
			subscriber: subscriber,
			publisher:  publisher,
			message:    message.NewMessage(stockManager),
			redisDB:    redisDB,
			mu:         mu,
		},
	)
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	serveCtx, cancelServeCtx := context.WithCancel(context.Background())
	defer cancelServeCtx()

	application := initApp()
	stockInit := stock.NewDataStockServer(application.publisher, application.redisDB, application.mu)
	if err := stockInit.LoadInitialData(serveCtx); err != nil {
		log.Println("error loading initial data", err.Error())
	}

	go stockInit.StartDirectoryMonitor(serveCtx, &wg)

	if err := application.Serve(serveCtx); err != nil {
		log.Println("failed to serve", err.Error())
	}
	defer func(application *App) {
		if err := application.Close(); err != nil {
			log.Println("error closing app", err.Error())
		}
	}(application)

	wg.Wait()
}

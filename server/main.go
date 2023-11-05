package main

import (
	"context"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/db"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/message"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net"
)

type App struct {
	gRPCServer *grpc.Server
	subscriber pubsub.Subscriber
	publisher  pubsub.Publisher
	message    *message.Message
	redisDB    *db.Redis
}

func NewApp(
	gRPCServer *grpc.Server,
	subscriber pubsub.Subscriber,
	publisher pubsub.Publisher,
	message *message.Message,
	redisDB *db.Redis,
) *App {
	app := &App{
		gRPCServer: gRPCServer,
		subscriber: subscriber,
		publisher:  publisher,
		message:    message,
		redisDB:    redisDB,
	}
	pb.RegisterDataStockServer(app.gRPCServer, stock.NewDataStockServer(app.publisher, app.redisDB))
	app.message.Load(app.subscriber)

	return app
}

func (a *App) Serve(ctx context.Context) error {
	if err := a.subscriber.Start(ctx); err != nil {
		return errors.Wrap(err, "error starting subscriber")
	}

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln("failed to listen", err.Error())
	}
	defer listen.Close()

	return a.gRPCServer.Serve(listen)
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

	dbRedis := db.NewRedis()
	dbRedisManager := db.NewRedisManager(dbRedis)
	stockInit := stock.NewDataStockServer(publisher, dbRedisManager)
	if err := stockInit.LoadInitialData(context.Background()); err != nil {
		log.Fatalln("error loading initial data", err.Error())
	}

	return NewApp(
		grpc.NewServer(),
		subscriber,
		publisher,
		message.NewMessage(),
		dbRedisManager,
	)
}

func main() {
	serveCtx, cancelServeCtx := context.WithCancel(context.Background())
	defer cancelServeCtx()
	if err := initApp().Serve(serveCtx); err != nil {
		log.Fatalln("failed to serve", err.Error())
	}
}

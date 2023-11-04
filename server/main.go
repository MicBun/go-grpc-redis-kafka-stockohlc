package main

import (
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/pubsub"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-server/stock/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type App struct {
	gRPCServer *grpc.Server
	subscriber pubsub.Subscriber
	publisher  pubsub.Publisher
}

func NewApp(
	gRPCServer *grpc.Server,
	subscriber pubsub.Subscriber,
	publisher pubsub.Publisher,
) *App {
	pb.RegisterDataStockServer(gRPCServer, stock.NewDataStockServer())
	if err := subscriber.Start(); err != nil {
		log.Fatalln("error starting subscriber", err.Error())
	}
	return &App{
		gRPCServer: gRPCServer,
		subscriber: subscriber,
		publisher:  publisher,
	}
}

func main() {
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalf("Server failed to load env file: %v", err)
		}
	}

	subscriber, err := pubsub.NewWatermillSubscriber()
	if err != nil {
		log.Fatalln("error creating watermill subscriber", err.Error())
	}

	publisher, err := pubsub.NewWatermillPublisher()
	if err != nil {
		log.Fatalln("error creating watermill publisher", err.Error())
	}

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln("failed to listen", err.Error())
	}

	log.Println("starting gRPC server at", listen.Addr().String())
	if err = NewApp(
		grpc.NewServer(),
		subscriber,
		publisher,
	).gRPCServer.Serve(listen); err != nil {
		log.Fatalln("failed to serve", err.Error())
	}
}

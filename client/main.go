package main

import (
	"context"
	"fmt"
	pb "github.com/MicBun/go-grpc-redis-kafka-stockohlc/stock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func getOneStockSummary(client pb.DataStockClient, stockSymbol string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stock, err := client.GetOneSummary(ctx, &pb.GetOneSummaryRequest{StockSymbol: stockSymbol})
	if err != nil {
		log.Fatalln("error getting stock summary", err.Error())
	}

	fmt.Println(stock)
}

func main() {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(":1200", opts...)
	if err != nil {
		log.Fatalln("fail to dial: ", err.Error())
	}
	defer conn.Close()

	client := pb.NewDataStockClient(conn)
	getOneStockSummary(client, "ASII")
	getOneStockSummary(client, "BBCA")
}

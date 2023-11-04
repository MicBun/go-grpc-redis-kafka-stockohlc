package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"time"
)

func getOneStockSummary(client pb.DataStockClient, stockSymbol string) (*pb.Stock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.GetOneSummary(ctx, &pb.GetOneSummaryRequest{StockSymbol: stockSymbol})
}

func main() {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial("grpcserver:50051", opts...)
	if err != nil {
		log.Fatalln("fail to dial: ", err.Error())
	}
	defer conn.Close()

	client := pb.NewDataStockClient(conn)

	r := mux.NewRouter()
	r.HandleFunc("/stock/{stockSymbol}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		stockSymbol := vars["stockSymbol"]
		stock, err := getOneStockSummary(client, stockSymbol)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stock)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	http.Handle("/", r)

	fmt.Println("Client HTTP server started on :8080")
	http.ListenAndServe(":8080", nil)
}

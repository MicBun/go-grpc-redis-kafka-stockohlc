package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	pb "github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	defer func(conn *grpc.ClientConn) {
		if err = conn.Close(); err != nil {
			log.Fatalln("fail to close: ", err.Error())
		}
	}(conn)

	r := mux.NewRouter()
	r.HandleFunc("/stock/{stockSymbol}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		stockSymbol := vars["stockSymbol"]
		stock, errGetOne := getOneStockSummary(pb.NewDataStockClient(conn), stockSymbol)
		if errGetOne != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(stock); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Println("Client HTTP server started on :8080")
	if err = server.ListenAndServe(); err != nil {
		panic(err)
	}
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock"
	"github.com/MicBun/go-grpc-redis-kafka-stockohlc-client/stock/pb"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

	stockClient := stock.NewManager()
	r := mux.NewRouter()
	r.HandleFunc("/stock/{stockSymbol}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		stockSymbol := vars["stockSymbol"]
		stockSummary, errGetOne := stockClient.GetOneStockSummary(pb.NewDataStockClient(conn), stockSymbol)
		if errGetOne != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(stockSummary); err != nil {
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

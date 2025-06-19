package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/logging"
)

var client benchclient.BenchClient

func orderbookHandler(w http.ResponseWriter, r *http.Request) {
	pair := r.URL.Query().Get("pair")
	if pair == "" {
		http.Error(w, "missing pair", http.StatusBadRequest)
		return
	}
	reply, err := client.ViewOrderbook(pair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply.Orderbook)
}

func pairsHandler(w http.ResponseWriter, r *http.Request) {
	reply, err := client.GetPairs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply.PairList)
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	pair := r.URL.Query().Get("pair")
	if pair == "" {
		http.Error(w, "missing pair", http.StatusBadRequest)
		return
	}
	reply, err := client.GetPrice(pair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply.Price)
}

func main() {
	var rpchost string
	var rpcport uint
	var webport uint

	flag.StringVar(&rpchost, "rpchost", "localhost", "RPC server host")
	flag.UintVar(&rpcport, "rpcport", 12345, "RPC server port")
	flag.UintVar(&webport, "webport", 8080, "web interface port")
	flag.Parse()

	if err := client.SetupBenchClient(rpchost, uint16(rpcport)); err != nil {
		logging.Fatalf("Error setting up RPC client: %v", err)
	}

	http.HandleFunc("/api/orderbook", orderbookHandler)
	http.HandleFunc("/api/pairs", pairsHandler)
	http.HandleFunc("/api/price", priceHandler)
	http.Handle("/", http.FileServer(http.Dir("cmd/webui/static")))

	addr := fmt.Sprintf(":%d", webport)
	logging.Infof("Web UI listening on %s", addr)
	logging.Fatal(http.ListenAndServe(addr, nil))
}

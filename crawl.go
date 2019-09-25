package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/dgraph-io/ristretto"
)

var (
	NWorkers = flag.Int("n", 4, "The number of workers to start")
	HTTPAddr = flag.String("http", "127.0.0.1:8000", "Address to listen for HTTP requests on")
)

var cache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: 1000000 * 10,
	MaxCost:     1000000,
	BufferItems: 64,
})

type safeVisited struct {
	v   *ristretto.Cache
	mux sync.Mutex
}

var visited = safeVisited{v: cache}

func main() {
	flag.Parse()

	// Start the dispatcher.
	fmt.Println("Starting the dispatcher")
	StartDispatcher(*NWorkers)

	// Register our collector as an HTTP handler function.
	fmt.Println("Registering the collector")
	http.HandleFunc("/uri", Collector)

	// Start the HTTP server!
	fmt.Println("HTTP server listening on", *HTTPAddr)
	if err := http.ListenAndServe(*HTTPAddr, nil); err != nil {
		fmt.Println(err.Error())
	}

}

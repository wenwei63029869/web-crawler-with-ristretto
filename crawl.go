package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/jackdanger/collectlinks"
)

type safeVisited struct {
	v   *ristretto.Cache
	mux sync.Mutex
}

func main() {
	var cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000000 * 10,
		MaxCost:     1000000,
		BufferItems: 64,
	})

	if err != nil {
		panic(err)
	}

	var visited = safeVisited{v: cache}

	flag.Parse()
	args := flag.Args()
	fmt.Println(args)

	if len(args) < 1 {
		fmt.Println("Please specify start page")
		os.Exit(1)
	}

	queue := make(chan string)

	go func() {
		queue <- args[0]
	}()

	for uri := range queue {
		retrieve(&visited, uri, queue)
	}
}

func retrieve(visited *safeVisited, uri string, queue chan string) {
	fmt.Println("fetching", uri)
	markVisited(visited, uri)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := http.Client{Transport: transport}

	resp, err := client.Get(uri)
	if err != nil {
		fmt.Println("http transport error is:", err)
		return
	}
	defer resp.Body.Close()

	links := collectlinks.All(resp.Body)

	enqueue(visited, uri, links, queue)
}

func markVisited(visited *safeVisited, uri string) {
	visited.mux.Lock()
	visited.v.Set(uri, true, 1)
	visited.mux.Unlock()
}

func enqueue(visited *safeVisited, uri string, links []string, queue chan string) {
	for _, link := range links {
		func() {
			visited.mux.Lock()
			absolute := strings.TrimRight(fixUrl(link, uri), "/")
			_, found := visited.v.Get(absolute)
			if absolute != "" && !found {
				go func() { queue <- absolute }()
			}
			visited.mux.Unlock()
		}()
	}
}

func fixUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}

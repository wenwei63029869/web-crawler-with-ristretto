package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackdanger/collectlinks"
)

func retrieve(visited *safeVisited, uri string) {
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

	enqueue(visited, uri, links)
}

func markVisited(visited *safeVisited, uri string) {
	visited.mux.Lock()
	visited.v.Set(uri, true, 1)
	visited.mux.Unlock()
}

func enqueue(visited *safeVisited, uri string, links []string) {
	for _, link := range links {
		func() {
			visited.mux.Lock()
			absolute := strings.TrimRight(fixURL(link, uri), "/")
			_, found := visited.v.Get(absolute)
			if absolute != "" && !found {
				WorkQueue <- absolute
			}
			visited.mux.Unlock()
		}()
	}
}

func fixURL(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseURL.ResolveReference(uri)
	return uri.String()
}

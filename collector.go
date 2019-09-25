package main

import (
	"fmt"
	"net/http"
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan string, 100)

func Collector(w http.ResponseWriter, r *http.Request) {
	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Now, we retrieve the uri from the request.
	uri := r.FormValue("uri")

	// Just do a quick bit of sanity checking to make sure the client actually provided us with a name.
	if uri == "" {
		http.Error(w, "You must specify a uri.", http.StatusBadRequest)
		return
	}

	// Now, we take the delay, and the person's name, and make a WorkRequest out of them.
	work := uri
	fmt.Println(uri)
	// Push the work onto the queue.
	WorkQueue <- work
	fmt.Println("Work request queued")

	// And let the user know their work request was created.
	w.WriteHeader(http.StatusCreated)
	return
}

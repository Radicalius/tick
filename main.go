package main

import (
	"fmt"
	"net/http"
	"tick/gen/tickv1connect"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const address = "localhost:8080"

func main() {
	mux := http.NewServeMux()
	path, handler := tickv1connect.NewTickHandler(&TickServiceServer{})
	mux.Handle(path, handler)
	fmt.Println("... Listening on", address)
	http.ListenAndServe(
		address,
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

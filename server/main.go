package main

import (
	"fmt"
	"log"
	"net/http"
	"tick/gen/tickv1connect"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const address = "0.0.0.0:8000"

func main() {
	if err := InitDatabase(); err != nil {
		log.Fatalf("error initializing the database: %s", err.Error())
	}

	mux := http.NewServeMux()
	path, handler := tickv1connect.NewTickHandler(&TickServiceServer{})
	mux.Handle(path, handler)
	fmt.Println("... Listening on", address)
	err := http.ListenAndServe(
		address,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("error starting server: %s", err.Error())
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"tick/gen/tickv1connect"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	if err := InitDatabase(); err != nil {
		log.Fatalf("error initializing the database: %s", err.Error())
	}

	go RunMetricsApi()

	mux := http.NewServeMux()
	path, handler := tickv1connect.NewTickHandler(&TickServiceServer{})
	mux.Handle(path, handler)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "8000"
	}

	address := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Println("... Listening on", address)

	err := http.ListenAndServe(
		address,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("error starting server: %s", err.Error())
	}
}

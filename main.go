package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"gopkg.in/macaroon-bakery.v0/bakery"
	"gopkg.in/macaroon-bakery.v0/bakery/checkers"
	"gopkg.in/macaroon-bakery.v0/httpbakery"
)

var (
	port = flag.String("port", ":8080", "port to serve one")
)

func checker(req *http.Request, cavId, cav string) ([]checkers.Caveat, error) {
	if cav != "is-timestamped" {
		return nil, fmt.Errorf("sorry")
	}
	return []checkers.Caveat{checkers.DeclaredCaveat("timestamp", time.Now().Format(time.RFC3339))}, nil
}

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", "localhost"+*port)
	if err != nil {
		panic(err)
	}
	endpointURL := "http://" + listener.Addr().String()
	p := bakery.NewServiceParams{
		Location: endpointURL,
	}
	service, err := bakery.NewService(p)
	if err != nil {
		panic(err)
	}
	fmt.Println("serving on " + endpointURL)
	mux := http.NewServeMux()
	httpbakery.AddDischargeHandler(mux, "/", service, checker)
	http.Serve(listener, mux)
}

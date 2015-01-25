// Package main contains an example client to test the timestamper
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"gopkg.in/macaroon-bakery.v0/bakery"
	"gopkg.in/macaroon-bakery.v0/bakery/checkers"
	"gopkg.in/macaroon-bakery.v0/httpbakery"
)

var (
	Url = flag.String("url", "http://127.0.0.1:8080", "url of the timestamper")
)

func noVisit(*url.URL) error {
	return fmt.Errorf("unexpected call to visit")
}

func getPublicKey(url string) (*bakery.PublicKey, error) { // This could be built in to the bakery.
	resp, err := http.Get(url + "/publickey")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var pubkey struct {
		PublicKey *bakery.PublicKey
	}
	err = decoder.Decode(&pubkey)
	if err != nil {
		return nil, err
	}
	return pubkey.PublicKey, nil
}

func main() {
	flag.Parse()
	var publicKey *bakery.PublicKey
	var err error
	locator := bakery.NewPublicKeyRing()
	publicKey, err = getPublicKey(*Url)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(publicKey)
	err = locator.AddPublicKeyForLocation(*Url, true, publicKey)
	if err != nil {
		log.Fatal(err)
	}
	svc, err := bakery.NewService(bakery.NewServiceParams{
		Locator: locator,
	})
	if err != nil {
		log.Fatal(err)
	}

	m, err := svc.NewMacaroon("", nil, []checkers.Caveat{{
		Location:  *Url,
		Condition: "is-timestamped",
	}})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("making request")

	client := httpbakery.NewHTTPClient()
	ms, err := httpbakery.DischargeAll(m, client, noVisit)
	if err != nil {
		log.Fatal(err)
	}
	declared := checkers.InferDeclared(ms)
	log.Println(declared["timestamp"])
	log.Println("done")
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	api "github.com/lifeboat008/lifeboat-api"
	ledger "github.com/lifeboat008/lifeboat-ledger"
)

func main() {
	var actors []api.Actor
	if err := json.Unmarshal([]byte(os.Getenv("LIFEBOAT_ACTORS_JSON")), &actors); err != nil {
		log.Fatal("LIFEBOAT_ACTORS_JSON must be a JSON array of actor id, role, and token")
	}
	secret := os.Getenv("LIFEBOAT_TESTNET_SECRET")
	if secret == "" {
		log.Fatal("LIFEBOAT_TESTNET_SECRET is required")
	}
	gateway, err := ledger.NewTestnetGateway(secret, nil)
	if err != nil {
		log.Fatal(err)
	}
	service, err := ledger.NewService(gateway)
	if err != nil {
		log.Fatal(err)
	}
	database := os.Getenv("LIFEBOAT_DB")
	if database == "" {
		database = "lifeboat.sqlite"
	}
	server, err := api.NewServer(database, actors, service)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	address := os.Getenv("LIFEBOAT_LISTEN")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	httpServer := &http.Server{Addr: address, Handler: server.Handler(), ReadHeaderTimeout: 10 * time.Second}
	log.Printf("Lifeboat API listening on %s", address)
	log.Fatal(httpServer.ListenAndServe())
}

package main

import (
	"log"
	"net/http"
	"os"

	"maskreview/internal/httpapi"
	"maskreview/internal/service"
	"maskreview/internal/store"
)

func main() {
	storePath := os.Getenv("MASKREVIEW_STORE_PATH")
	if storePath == "" {
		storePath = "./data/store.json"
	}

	st, err := store.NewFileStore(storePath)
	if err != nil {
		log.Fatalf("failed to init store: %v", err)
	}

	svc := service.NewService(st)

	handler := httpapi.NewHandler(svc)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

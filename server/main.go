package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
	_ "github.com/lib/pq"
)

func main() {
	cfg := LoadConfigFromEnv()
	cache := NewLRUCache(cfg.CacheCapacity)

	db, err := InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
	log.Fatal("Failed creating schema:", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Minute)


	router := mux.NewRouter()
	router.HandleFunc("/kv/{key}", handlePut(cache, db)).Methods("PUT")
	router.HandleFunc("/kv/{key}", handleGet(cache, db)).Methods("GET")
	router.HandleFunc("/kv/{key}", handleDelete(cache, db)).Methods("DELETE")
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(getMetrics())
	}).Methods("GET")

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Println("Server running at", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}


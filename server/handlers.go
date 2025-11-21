package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func handlePut(c *LRUCache, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		key := mux.Vars(r)["key"]
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.Exec("INSERT INTO kv VALUES ($1,$2) ON CONFLICT(key) DO UPDATE SET value=$2", key, body["value"])
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		c.Put(key, body["value"])
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

		record(time.Since(start))
	}
}

func handleGet(c *LRUCache, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		key := mux.Vars(r)["key"]

		if v, found := c.Get(key); found {
			json.NewEncoder(w).Encode(map[string]string{"value": v})
			record(time.Since(start))
			return
		}

		var value string
		err := db.QueryRow("SELECT value FROM kv WHERE key=$1", key).Scan(&value)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}

		c.Put(key, value)
		json.NewEncoder(w).Encode(map[string]string{"value": value})

		record(time.Since(start))
	}
}

func handleDelete(c *LRUCache, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		key := mux.Vars(r)["key"]
		_, err := db.Exec("DELETE FROM kv WHERE key=$1", key)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		c.Delete(key)
		json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

		record(time.Since(start))
	}
}


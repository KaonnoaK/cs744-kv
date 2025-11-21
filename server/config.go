package main

import (
	"os"
	"strconv"
)

type Config struct {
	DbURL         string
	Port          int
	CacheCapacity int
}

func LoadConfigFromEnv() Config {
	port := 8080
	if p := os.Getenv("KV_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	cacheCap := 10000
	if c := os.Getenv("KV_CACHE_CAP"); c != "" {
		if v, err := strconv.Atoi(c); err == nil {
			cacheCap = v
		}
	}
	
	dbURL := "postgres://kvuser:12345@localhost:5432/kvdb?sslmode=disable"
	if dbURL == "" {
		dbURL = "postgres://12345:kvpass@db:5432/kvdb?sslmode=disable"
	}

	return Config{DbURL: dbURL, Port: port, CacheCapacity: cacheCap}
}


//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Port     int    `json:"port"`
	Host     string `json:"host"`
	LogLevel string `json:"log_level"`
}

func main() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	fmt.Printf("stdlib: %+v\n", cfg)
}

//go:build ignore

package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func main() {
	k := koanf.New(".")
	if err := k.Load(file.Provider("config.json"), json.Parser()); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("koanf: port=%d host=%s\n", k.Int("port"), k.String("host"))
}

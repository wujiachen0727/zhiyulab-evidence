//go:build ignore

package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.json")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	fmt.Printf("viper: port=%d host=%s\n", viper.GetInt("port"), viper.GetString("host"))
}

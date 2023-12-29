package main

import (
	"log"

	"github.com/strider2038/key-value-database/internal/config"
)

func main() {
	params, err := config.Load()
	if err != nil {
		log.Fatalln("load config:", err)
	}

	log.Println(params)
}

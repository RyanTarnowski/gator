package main

import (
	"fmt"
	"gator/internal/config"
	"log"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	err = cfg.SetUser("Ryan")
	if err != nil {
		log.Fatalf("error setting username: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	fmt.Printf("%+v\n", cfg)
}

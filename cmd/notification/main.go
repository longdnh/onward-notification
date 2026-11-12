package main

import (
	"log"

	"github.com/longdnh/onward-notification/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("rabbit=%s exchange=%s queue=%s prefetch=%d", cfg.RabbitURL, cfg.Exchange, cfg.Queue, cfg.Prefetch)
}

package main

import (
	"log"

	"github.com/longdnh/onward-notification/internal/config"
	"github.com/longdnh/onward-notification/internal/rabbit"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("rabbit=%s exchange=%s queue=%s prefetch=%d", cfg.RabbitURL, cfg.Exchange, cfg.Queue, cfg.Prefetch)

	if err := rabbit.Run(cfg); err != nil {
		log.Fatalf("rabbit: %v", err)
	}
}

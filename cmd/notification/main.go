package main

import (
	"log"

	"github.com/longdnh/onward-notification/internal/config"
	"github.com/longdnh/onward-notification/internal/dedup"
	"github.com/longdnh/onward-notification/internal/mail"
	"github.com/longdnh/onward-notification/internal/rabbit"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	log.Printf("rabbit=%s exchange=%s queue=%s prefetch=%d", cfg.RabbitURL, cfg.Exchange, cfg.Queue, cfg.Prefetch)

	composer := mail.NewComposer(cfg.FrontendBaseURL)
	sender := mail.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.MailFrom)

	handler := mail.NewHandler(composer, sender)
	dedupIns := dedup.NewRedis(cfg.RedisAddr, cfg.DedupTTL)

	if err := rabbit.Run(cfg, handler, dedupIns); err != nil {
		log.Fatalf("rabbit: %v", err)
	}
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RabbitURL string // amqp://user:pass@host:port/
	Exchange  string // owned and declared by Account; the contract between the two services
	Queue     string // owned and declared by this service
	Prefetch  int    // max unacked messages delivered to this consumer at once

	SMTPHost string
	SMTPPort int
	MailFrom string

	FrontendBaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		RabbitURL:       getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange:        getenv("RABBITMQ_EXCHANGE", "onward.auth"),
		Queue:           getenv("RABBITMQ_QUEUE", "notification.auth-mail"),
		SMTPHost:        getenv("SMTP_HOST", "localhost"),
		MailFrom:        getenv("MAIL_FROM", "no-reply@onward.local"),
		FrontendBaseURL: strings.TrimSuffix(getenv("FRONTEND_BASE_URL", "http://localhost:5173"), "/"),
	}

	var err error
	if cfg.Prefetch, err = getenvInt("RABBITMQ_PREFETCH", 10); err != nil {
		return Config{}, err
	}
	if cfg.SMTPPort, err = getenvInt("SMTP_PORT", 1025); err != nil {
		return Config{}, err
	}

	if cfg.Prefetch < 1 {
		return Config{}, fmt.Errorf("RABBITMQ_PREFETCH must be >= 1, got %d", cfg.Prefetch)
	}
	if !strings.HasPrefix(cfg.FrontendBaseURL, "http://") && !strings.HasPrefix(cfg.FrontendBaseURL, "https://") {
		return Config{}, fmt.Errorf("FRONTEND_BASE_URL must be an absolute http(s) URL, got %q", cfg.FrontendBaseURL)
	}
	return cfg, nil
}

// getenv returns the variable's value, or def when it is unset or empty.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q", key, v)
	}
	return n, nil
}

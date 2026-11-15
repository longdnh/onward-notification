package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

	RedisAddr string        // host:port of the Redis that remembers processed message IDs
	DedupTTL  time.Duration // how long a processed ID is remembered; must exceed the longest token lifetime
}

func Load() (Config, error) {
	cfg := Config{
		RabbitURL:       getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange:        getenv("RABBITMQ_EXCHANGE", "onward.auth"),
		Queue:           getenv("RABBITMQ_QUEUE", "notification.auth-mail"),
		SMTPHost:        getenv("SMTP_HOST", "localhost"),
		MailFrom:        getenv("MAIL_FROM", "no-reply@onward.local"),
		FrontendBaseURL: strings.TrimSuffix(getenv("FRONTEND_BASE_URL", "http://localhost:5173"), "/"),
		RedisAddr:       getenv("REDIS_ADDR", "localhost:6379"),
	}

	var err error
	if cfg.Prefetch, err = getenvInt("RABBITMQ_PREFETCH", 10); err != nil {
		return Config{}, err
	}
	if cfg.SMTPPort, err = getenvInt("SMTP_PORT", 1025); err != nil {
		return Config{}, err
	}
	if cfg.DedupTTL, err = getenvDuration("DEDUP_TTL", 48*time.Hour); err != nil {
		return Config{}, err
	}

	if cfg.DedupTTL < time.Minute {
		return Config{}, fmt.Errorf("DEDUP_TTL must be at least 1m, got %s", cfg.DedupTTL)
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

// getenvDuration parses values like "48h", "30m" or "1h30m".
func getenvDuration(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration like 48h, got %q", key, v)
	}
	return d, nil
}

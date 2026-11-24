# onward-notification

Consumer that sends authentication mails for the Onward system. It receives
events from RabbitMQ (published by the Account service), sends mail over SMTP,
and uses Redis to avoid sending duplicates when a message is redelivered.

It handles two event types:

| `type` | Mail sent | Link in the mail |
|---|---|---|
| `onward.auth.email-verification-requested.v1` | Confirm your email address | `<FRONTEND_BASE_URL>/verify-email?token=...` |
| `onward.auth.password-reset-requested.v1` | Reset your password | `<FRONTEND_BASE_URL>/reset-password?token=...` |

## Requirements

- [Nix](https://nixos.org/download) (`shell.nix` provides Go 1.27 and gopls).
- [direnv](https://direnv.net/) (optional, loads the environment automatically when you `cd` into the directory).
- Docker (to run RabbitMQ, Redis and Mailpit locally).

## Running locally

### 1. Enter the Nix environment

```bash
git clone <repo-url>
cd onward-notification

# Option 1: direnv, allow once
direnv allow

# Option 2: manually, in every new terminal
nix-shell
```

VS Code: install the **Nix Environment Selector** extension. The workspace
already points it at `shell.nix` in `.vscode/settings.json`.

### 2. Create the environment file

```bash
cp .env.example .env
```

The defaults in `.env.example` match the containers in step 3, so nothing
needs to change for local development. The application does not read `.env`
itself: direnv loads it through `.envrc`. Without direnv:

```bash
set -a; source .env; set +a
```

### 3. Start the dependencies

```bash
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
docker run -d --name redis    -p 6379:6379 redis:7
docker run -d --name mailpit  -p 1025:1025 -p 8025:8025 axllent/mailpit
```

| Service | Port | UI |
|---|---|---|
| RabbitMQ | 5672 | http://localhost:15672 (guest / guest) |
| Redis | 6379 | |
| Mailpit (fake SMTP) | 1025 | http://localhost:8025 |

The `onward.auth` exchange is declared by the Account service. If you run this
service on its own, create that exchange first in the RabbitMQ UI
(Exchanges > Add a new exchange, type `topic`, durable).

### 4. Run the service

```bash
go run ./cmd/notification
```

On startup the service declares the `notification.auth-mail` queue (quorum),
binds it to the exchange with routing key `auth.#`, and declares the
dead-letter exchange `onward.auth.dlx` and queue `notification.auth-mail.dlq`.

### 5. Send a test message

In the RabbitMQ UI, open the `onward.auth` exchange > Publish message, routing
key `auth.test`, payload:

```json
{
  "messageId": "test-1",
  "type": "onward.auth.email-verification-requested.v1",
  "occurredAt": "2026-11-24T10:00:00Z",
  "email": "user@example.com",
  "token": "abc123",
  "expiresAt": "2026-11-25T10:00:00Z"
}
```

The mail shows up in Mailpit at http://localhost:8025. Publishing again with
the same `messageId` is skipped (logged as `Duplicated Message`).

## Configuration

Everything is configured through environment variables, and every variable
has a default. See `.env.example`.

| Variable | Default | Notes |
|---|---|---|
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | |
| `RABBITMQ_EXCHANGE` | `onward.auth` | Owned by the Account service |
| `RABBITMQ_QUEUE` | `notification.auth-mail` | Declared by this service |
| `RABBITMQ_PREFETCH` | `10` | >= 1 |
| `SMTP_HOST` | `localhost` | |
| `SMTP_PORT` | `1025` | |
| `MAIL_FROM` | `no-reply@onward.local` | |
| `FRONTEND_BASE_URL` | `http://localhost:5173` | Must be an absolute http(s) URL |
| `REDIS_ADDR` | `localhost:6379` | |
| `DEDUP_TTL` | `48h` | >= 1m, must exceed the longest token lifetime |

SMTP currently sends without auth or TLS, which is enough for Mailpit. A real
provider needs auth and STARTTLS added in `internal/mail/smtp.go`.

## Message processing

1. Decode and validate the JSON. Malformed messages are nacked without
   requeue and land in the dead-letter queue.
2. `SETNX` the key `notification:processed:<messageId>` in Redis with TTL
   `DEDUP_TTL`. If the key already exists, ack and skip.
3. Send the mail. On failure, delete the Redis key so the redelivery is not
   treated as a duplicate, and the service exits with an error (the unacked
   message is redelivered by RabbitMQ).
4. On success, ack.

The service shuts down gracefully on `SIGINT` or `SIGTERM`.

## Project layout

```
cmd/notification/   entrypoint
internal/config/    reads and validates environment variables
internal/event/     TokenMail struct, decode + validate
internal/rabbit/    queue declaration, consume, dedup, ack/nack
internal/mail/      composes mail content and sends it over SMTP
internal/dedup/     duplicate protection with Redis SETNX
```

## Build and Docker

```bash
go build -o bin/notification ./cmd/notification
go vet ./...

docker build -t onward-notification .
docker run --rm --env-file .env --network host onward-notification
```

The image is distroless, runs as a non-root user, and contains only the
static binary.

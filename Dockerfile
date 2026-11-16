# Stage 1: build a static binary. CGO_ENABLED=0 removes the libc dependency so
# the result runs in an image with no OS libraries at all.
FROM golang:1.27-alpine AS build
WORKDIR /src

# Dependencies first, so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/notification ./cmd/notification

# Stage 2: only the binary, no shell, no package manager, non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/notification /notification
USER nonroot
ENTRYPOINT ["/notification"]

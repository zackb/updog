FROM golang:1.24-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache build-base musl-dev

WORKDIR /build

# Copy Go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source code
COPY . .

# Build static Go binary with embedded frontend
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -v -tags "sqlite_omit_load_extension" \
        -ldflags '-linkmode external -extldflags "-static"' \
        -o updog ./cmd/updog/updog.go

# Final minimal image
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the static binary
COPY --from=backend-builder /build/updog /updog/updog

# Create tmp directory for SQLite
COPY --from=backend-builder /tmp /tmp

WORKDIR /updog

EXPOSE 8080

CMD ["./updog"]


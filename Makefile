OUT=updog

default: build

run: build
	./$(OUT)

run-dev: build
	DEV=1 ./$(OUT)

build:
	go build -v -o $(OUT) ./cmd/$(OUT)/$(OUT).go

build-static:
	CGO_ENABLED=0 GOOS=linux go build $(TAGS) -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(OUT) -v ./cmd/$(OUT)/$(OUT).go

build-static-musl:
	CGO_ENABLED=1 GOOS=linux CC="musl-gcc" \
    go build -tags "sqlite_omit_load_extension" \
    -ldflags '-linkmode external -extldflags "-static"' \
    -o $(OUT) ./cmd/$(OUT)/$(OUT).go

docker-setup:
	@echo "Setting up Docker buildx for multi-architecture builds..."
	@docker buildx inspect $(OUT)-builder >/dev/null 2>&1 || \
		(echo "Creating new buildx builder..." && \
		 docker buildx create --name $(OUT)-builder --bootstrap)
	@docker buildx use $(OUT)-builder
	@echo "✅ Docker buildx builder ready: $(OUT)-builder"

docker-local:
	docker build -f Dockerfile -t $(OUT)/$(OUT):latest .
	@echo "Local Docker image built: $(OUT)/$(OUT):latest"

docker-multiarch: docker-setup
	@echo "Building multi-architecture Docker images..."
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--cache-to=type=local,dest=.buildx-cache \
		--cache-from=type=local,src=.buildx-cache \
		--tag registry.bartel.com/$(OUT)/$(OUT):latest \
		--tag registry.bartel.com/$(OUT)/$(OUT):$$(date +%Y%m%d) \
		--push \
		.
	@echo "Multi-arch images pushed: $(OUT)/$(OUT):latest"

docker: docker-local

jwk-key-dev:
	go run cmd/jwk/jwk.go -c > jwks.json

test:
	go test ./...

updatedeps:
	go list -m -u all

cleandeps:
	go mod tidy

clean:
	go clean
	go clean --modcache

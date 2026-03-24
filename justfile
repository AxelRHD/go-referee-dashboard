app_name := "referee-dashboard"
bin_dir := "bin"
bin_file := bin_dir / app_name
version := `git describe --tags --always 2>/dev/null || echo "dev"`

[private]
default:
    @just --list --unsorted

# ============================================================
# Development
# ============================================================

# Start dev server with hot-reload
[group('dev')]
dev:
    @air

# Format code
[group('dev')]
fmt:
    @go fmt ./...

# Static analysis
[group('dev')]
vet:
    @go vet ./...

# Run tests
[group('dev')]
test:
    @go test -v ./...

# ============================================================
# Build
# ============================================================

# Build binary
[group('build')]
build:
    @mkdir -p {{bin_dir}}
    @CGO_ENABLED=0 go build -ldflags "-X main.version={{version}} -s -w" -o {{bin_file}} ./cmd

# ============================================================
# Database
# ============================================================

# Run database migrations
[group('db')]
migrate:
    @go run ./cmd migrate

# Seed initial data (positions)
[group('db')]
seed:
    @go run ./cmd seed

# Generate sqlc code
[group('db')]
generate:
    @sqlc generate

# ============================================================
# Deployment
# ============================================================

remote := "mimir"
appdata_dir := "/mnt/data/docker/appdata/referee-dashboard"

# Build Docker image (uses local binary)
[group('deploy')]
build-image: build
    docker build -t {{app_name}} .

# Push image to server and prune old images
[group('deploy')]
deploy-image:
    docker save {{app_name}} | ssh {{remote}} docker load
    ssh {{remote}} docker image prune -f

# Full deploy: build + push image
[group('deploy')]
deploy: build-image deploy-image

# Show container logs
[group('deploy')]
deploy-logs:
    gosctl run logs

# Show container status
[group('deploy')]
deploy-status:
    gosctl run status

# ============================================================
# Clean
# ============================================================

# Remove build artifacts
[group('clean')]
clean:
    @rm -rf {{bin_dir}}

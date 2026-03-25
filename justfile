set dotenv-load

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

# Generate sqlc code
[group('db')]
generate:
    @sqlc generate

# ============================================================
# Seeding
# ============================================================

seed_dir := "db/seed"
db_file := env("DB_PATH", "referee.db")

# Export current positions, leagues, teams, venues as seed SQL (without IDs/timestamps)
[group('seeding')]
dump-seed:
    @sqlite3 {{db_file}} ".mode insert positions" ".output {{seed_dir}}/positions.sql" "SELECT position, long, sorter FROM positions ORDER BY sorter;"
    @sqlite3 {{db_file}} ".mode insert leagues" ".output {{seed_dir}}/leagues.sql" "SELECT name, short_name, sorter, remarks FROM leagues ORDER BY sorter, name;"
    @sqlite3 {{db_file}} ".mode insert teams" ".output {{seed_dir}}/teams.sql" "SELECT name, state, is_active, remarks FROM teams ORDER BY name;"
    @sqlite3 {{db_file}} ".mode insert venues" ".output {{seed_dir}}/venues.sql" "SELECT city, short_name, stadium, lat, lon FROM venues WHERE id > 0 ORDER BY short_name, city;"
    @echo "Seed data exported to {{seed_dir}}/"

# Import seed data (positions, leagues, teams, venues)
[group('seeding')]
[confirm("This will import seed data into the database. Continue?")]
seed-data:
    @sqlite3 {{db_file}} < {{seed_dir}}/positions.sql
    @sqlite3 {{db_file}} < {{seed_dir}}/leagues.sql
    @sqlite3 {{db_file}} < {{seed_dir}}/teams.sql
    @sqlite3 {{db_file}} < {{seed_dir}}/venues.sql
    @echo "Seed data imported."

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

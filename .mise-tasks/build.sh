#!/usr/bin/env bash
#MISE description="Binary bauen (mit Git-Version)"
set -euo pipefail
ver=$(git describe --tags --always 2>/dev/null || echo "dev")
mkdir -p "$BIN_DIR"
CGO_ENABLED=0 go build -ldflags "-X main.version=${ver} -s -w" -o "$BIN_DIR/$APP_NAME" ./cmd

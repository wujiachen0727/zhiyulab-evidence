#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$(cd "$SCRIPT_DIR/../../output/error-boundary-compare" && pwd)"

{
  printf '[实测环境]\n'
  go version
  docker run --rm php:8.4-cli php -v | head -n 1
  printf '\n[PHP weak types]\n'
  docker run --rm -v "$SCRIPT_DIR:/work" -w /work php:8.4-cli php php_weak.php
  printf '\n[PHP strict_types=1]\n'
  docker run --rm -v "$SCRIPT_DIR:/work" -w /work php:8.4-cli php php_strict.php
  printf '\n[Go JSON struct decode]\n'
  go run "$SCRIPT_DIR/main.go"
} | tee "$OUTPUT_DIR/result.txt"

#!/usr/bin/env bash
set -euo pipefail

version="v2.3.0"

if command -v golangci-lint >/dev/null 2>&1 && golangci-lint version 2>/dev/null | grep -q 'version 2\.'; then
  exit 0
fi

bin_dir="/usr/local/bin"
if [ ! -w "$bin_dir" ]; then
  bin_dir="$HOME/.local/bin"
  mkdir -p "$bin_dir"
fi

curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$bin_dir" "$version"

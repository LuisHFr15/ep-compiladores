#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go build -o "$tmp/cobolc" ./cmd/cobolc || exit 1

total=0
passed=0

for file in tests/*.cob; do
	total=$((total + 1))
	echo "=== $file ==="
	if "$tmp/cobolc" "$file" 2>&1; then
		passed=$((passed + 1))
		echo "[OK]"
	else
		echo "[FALHA]"
	fi
	echo
done

echo "Resumo: $passed/$total passaram"
[ "$passed" -eq "$total" ]

#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
inputs="tests/input_programs"
outputs="tests/outputs"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go build -o "$tmp/cobolc" ./cmd/cobolc || exit 1

mkdir -p "$outputs"
rm -f "$outputs"/*.txt

total=0
passed=0

for file in "$inputs"/*.cob; do
	total=$((total + 1))
	out="$outputs/$(basename "$file" .cob).txt"
	echo "=== $file -> $out ==="
	if "$tmp/cobolc" -o "$out" "$file" 2>&1; then
		passed=$((passed + 1))
		echo "[OK]"
	else
		echo "[FALHA]"
	fi
	echo
done

echo "Resumo: $passed/$total passaram"
[ "$passed" -eq "$total" ]

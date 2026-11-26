#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
outputs="tests/outputs"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go build -o "$tmp/cobolc" ./cmd/cobolc || exit 1

mkdir -p "$outputs"
rm -f "$outputs"/*.txt

total=0
passed=0

while read -r name _; do
	[ -z "$name" ] && continue
	total=$((total + 1))
	file="$(ls tests/input_programs/"$name".cob tests/input_programs/"$name".cbl tests/feature_programs/"$name".cob 2>/dev/null | head -1)"
	out="$outputs/$name.txt"
	echo "=== $file -> $out ==="
	if "$tmp/cobolc" -o "$out" -asm "$tmp/$name.asm" "$file" > /dev/null 2>&1; then
		passed=$((passed + 1))
		echo "[OK]"
	else
		echo "[FALHA]"
	fi
	echo
done < tests/programs.txt

echo "Resumo: $passed/$total passaram"
[ "$passed" -eq "$total" ]

#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
inputs="tests/error_programs"
outputs="tests/error_outputs"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go build -o "$tmp/cobolc" ./cmd/cobolc || exit 1

mkdir -p "$outputs"
rm -f "$outputs"/*.txt

total=0
detected=0
errors=0

for file in "$inputs"/*.cob; do
	total=$((total + 1))
	out="$outputs/$(basename "$file" .cob).txt"
	echo "=== $file -> $out ==="
	if "$tmp/cobolc" -o "$out" "$file" >/dev/null 2>&1; then
		echo "[FALHA] nenhum erro detectado"
	else
		detected=$((detected + 1))
		n=$(grep -oE 'Erros léxicos \([0-9]+\)' "$out" | grep -oE '[0-9]+')
		errors=$((errors + ${n:-0}))
		grep -E '^\[[IWES]\]' "$out"
		echo "[OK] $n erro(s) detectado(s)"
	fi
	echo
done

echo "Resumo: $detected/$total programas com erros detectados, $errors erros no total"
[ "$detected" -eq "$total" ]

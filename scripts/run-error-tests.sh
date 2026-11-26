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

rej_total=0
rej_ok=0
while read -r name expected; do
	[ -z "$name" ] && continue
	rej_total=$((rej_total + 1))
	file="$(ls tests/input_programs/"$name".cob tests/input_programs/"$name".cbl tests/semantic_programs/"$name".cob 2>/dev/null | head -1)"
	out="$outputs/${name}_rejeitado.txt"
	if "$tmp/cobolc" -o "$out" -asm "$tmp/$name.asm" "$file" >/dev/null 2>&1; then
		echo "[FALHA] $name: foi aceito"
	elif grep -q -- "$expected" "$out"; then
		rej_ok=$((rej_ok + 1))
		echo "[OK] $name rejeitado: $expected"
	else
		echo "[FALHA] $name: rejeitado sem a mensagem esperada ($expected)"
	fi
done < tests/rejected.txt

echo "Rejeitados: $rej_ok/$rej_total"

acc_total=0
acc_ok=0
while read -r name _; do
	[ -z "$name" ] && continue
	acc_total=$((acc_total + 1))
	file="tests/lexical_programs/$name.cob"
	out="$outputs/${name}_aceito.txt"
	if "$tmp/cobolc" -o "$out" -asm "$tmp/$name.asm" "$file" >/dev/null 2>&1; then
		acc_ok=$((acc_ok + 1))
		echo "[OK] $name aceito"
	else
		echo "[FALHA] $name: deveria ser aceito"
		grep -E '^\[E\]' "$out" | head -3
	fi
done < tests/accepted.txt

echo "Aceitos: $acc_ok/$acc_total"
[ "$detected" -eq "$total" ] && [ "$rej_ok" -eq "$rej_total" ] && [ "$acc_ok" -eq "$acc_total" ]

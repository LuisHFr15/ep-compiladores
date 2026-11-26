#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mars="$root/tools/mars/Mars4_5.jar"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go build -o "$tmp/cobolc" ./cmd/cobolc || exit 1
mkdir -p tests/asm

only="${1:-}"
total=0
passed=0

while read -r name opts; do
	[ -z "$name" ] && continue
	[ -n "$only" ] && [ "$only" != "$name" ] && continue
	total=$((total + 1))
	src="$(ls tests/input_programs/"$name".cob tests/input_programs/"$name".cbl 2>/dev/null | head -1)"
	asm="tests/asm/$name.asm"

	if ! "$tmp/cobolc" -asm "$asm" "$src" > "$tmp/$name.compile" 2>&1; then
		echo "[FALHA] $name: erro de compilacao"
		grep -E '^\[E\]|^erro' "$tmp/$name.compile" | head -5
		continue
	fi

	work="$tmp/work_$name"
	mkdir -p "$work"
	[ -d "tests/data/$name" ] && cp -R "tests/data/$name/." "$work/"
	(cd "$work" && ls -A) > "$tmp/$name.before"

	ok=1
	for suffix in "" 2; do
		in=/dev/null
		[ -f "tests/stdin/$name.in$suffix" ] && in="$root/tests/stdin/$name.in$suffix"
		[ -n "$suffix" ] && [ "$in" = /dev/null ] && continue
		(cd "$work" && perl -e 'alarm shift; exec @ARGV' 120 java -Dfile.encoding=ISO-8859-1 -Dstdout.encoding=ISO-8859-1 -jar "$mars" nc me sm "$root/$asm" < "$in" > "$tmp/$name.out$suffix" 2> "$tmp/$name.err$suffix")
		if ! cmp -s "$tmp/$name.out$suffix" "tests/expected/$name.out$suffix"; then
			ok=0
			echo "[FALHA] $name: saida$suffix difere"
			diff <(cat -v "tests/expected/$name.out$suffix") <(cat -v "$tmp/$name.out$suffix") | head -8
			head -3 "$tmp/$name.err$suffix"
		fi
	done

	if [ "$opts" != "nofiles" ] && [ -d "tests/expected/$name.files" ]; then
		for f in "tests/expected/$name.files"/*; do
			if ! cmp -s "$f" "$work/$(basename "$f")"; then
				ok=0
				echo "[FALHA] $name: arquivo $(basename "$f") difere"
			fi
		done
	fi

	if [ "$ok" -eq 1 ]; then
		passed=$((passed + 1))
		echo "[OK] $name"
	fi
done < tests/programs.txt

echo
echo "Resumo: $passed/$total programas com saida identica ao GnuCOBOL"
[ "$passed" -eq "$total" ]

#!/usr/bin/env bash
set -u

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cd "$root"
rm -rf tests/expected
mkdir -p tests/expected
printf '!*.out\n' > tests/expected/.gitignore

while read -r name opts; do
	[ -z "$name" ] && continue
	src="$(ls tests/input_programs/"$name".cob tests/input_programs/"$name".cbl tests/feature_programs/"$name".cob 2>/dev/null | head -1)"
	work="$tmp/$name"
	mkdir -p "$work"
	[ -d "tests/data/$name" ] && cp -R "tests/data/$name/." "$work/"
	(cd "$work" && ls -A) > "$tmp/$name.before"

	srcs="$src"
	[ -f "tests/drivers/$name.cob" ] && srcs="tests/drivers/$name.cob $src"
	if ! cobc -x -std=ibm -I tests/input_programs -o "$tmp/$name.bin" $srcs 2> "$tmp/$name.cobc.log"; then
		echo "[FALHA] cobc nao compilou $src"
		cat "$tmp/$name.cobc.log"
		continue
	fi

	for suffix in "" 2; do
		in=/dev/null
		[ -f "tests/stdin/$name.in$suffix" ] && in="$root/tests/stdin/$name.in$suffix"
		[ -n "$suffix" ] && [ "$in" = /dev/null ] && continue
		(cd "$work" && "$tmp/$name.bin" < "$in" > "$root/tests/expected/$name.out$suffix" 2> "$tmp/$name.err")
		echo "[OK] $name$suffix (exit $?)"
	done

	if [ "$opts" != "nofiles" ]; then
		for f in $(cd "$work" && ls -A); do
			grep -qxF "$f" "$tmp/$name.before" && continue
			mkdir -p "tests/expected/$name.files"
			cp "$work/$f" "tests/expected/$name.files/$f"
		done
	fi
done < tests/programs.txt

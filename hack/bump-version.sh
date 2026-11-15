#!/bin/sh
# Bump VERSION using the YY.MM.BuildNr scheme.
# BuildNr increments within a month and resets to 1 on the first day of a new
# month. The file is the single source of truth for release.yaml.
set -eu

file="${1:-VERSION}"
current="$(cat "$file" 2>/dev/null || echo '')"
now="$(date -u +%y.%m)"

build=1
if [ -n "$current" ]; then
  case "$current" in
    "$now".*)
      n="${current##*.}"
      case "$n" in
        ''|*[!0-9]*) echo "invalid VERSION: $current" >&2; exit 1 ;;
      esac
      build=$((n + 1))
      ;;
    *)
      build=1
      ;;
  esac
fi

next="$now.$build"
printf '%s\n' "$next" > "$file"
printf 'VERSION %s -> %s\n' "${current:-<none>}" "$next"

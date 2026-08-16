#!/bin/sh
set -eu

APP_PATH=""
if [ -d "/Applications/WebRecap.app" ]; then
  APP_PATH="/Applications/WebRecap.app"
elif [ -d "$HOME/Applications/WebRecap.app" ]; then
  APP_PATH="$HOME/Applications/WebRecap.app"
else
  echo "WebRecap.app not found in /Applications or ~/Applications" >&2
  exit 1
fi

if [ "$#" -eq 0 ]; then
  cat >&2 <<'EOF'
Usage:
  web-recap-safari bookmarks --browser safari
  web-recap-safari tabs --browser safari
  web-recap-safari --browser safari --date "$(date +%F)"
  web-recap-safari bookmarks --browser safari -o output.json

Notes:
  - Launches the FDA-enabled WebRecap.app via macOS 'open'
  - If you do not pass -o/--output, output is captured to a temp file and printed
  - Relative output paths are converted to absolute paths before launching the app
  - Intended for Safari commands where direct CLI execution is blocked by macOS privacy
EOF
  exit 1
fi

abs_path() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s/%s\n' "$PWD" "$1" ;;
  esac
}

has_output=0
output_path=""
next_is_output_path=0
normalized_args=""
for arg in "$@"; do
  if [ "$next_is_output_path" -eq 1 ]; then
    has_output=1
    output_path=$(abs_path "$arg")
    normalized_args="$normalized_args
$output_path"
    next_is_output_path=0
    continue
  fi

  case "$arg" in
    -o|--output)
      normalized_args="$normalized_args
$arg"
      next_is_output_path=1
      ;;
    --output=*)
      has_output=1
      output_path=$(abs_path "${arg#*=}")
      normalized_args="$normalized_args
--output=$output_path"
      ;;
    *)
      normalized_args="$normalized_args
$arg"
      ;;
  esac
done

if [ "$next_is_output_path" -eq 1 ]; then
  echo "Missing value for -o/--output" >&2
  exit 1
fi

cleanup_tmp=0
if [ "$has_output" -eq 0 ]; then
  output_path=$(mktemp /tmp/web-recap-safari.XXXXXX.json)
  cleanup_tmp=1
  normalized_args="$normalized_args
-o
$output_path"
fi

cleanup() {
  if [ "$cleanup_tmp" -eq 1 ]; then
    rm -f "$output_path"
  fi
}
trap cleanup EXIT INT TERM

rm -f "$output_path"

set --
IFS='
'
for arg in $normalized_args; do
  [ -n "$arg" ] || continue
  set -- "$@" "$arg"
done
unset IFS

open -n -a "$APP_PATH" --args "$@" >/dev/null 2>/dev/null

found=0
prev_size=-1
stable_count=0
attempt=0
while [ "$attempt" -lt 120 ]; do
  if [ -f "$output_path" ]; then
    found=1
    size=$(wc -c < "$output_path" | tr -d ' ')
    if [ "$size" = "$prev_size" ] && [ "$size" -gt 0 ] 2>/dev/null; then
      stable_count=$((stable_count + 1))
      if [ "$stable_count" -ge 2 ]; then
        break
      fi
    else
      stable_count=0
      prev_size=$size
    fi
  fi
  attempt=$((attempt + 1))
  sleep 0.25
done

if [ "$found" -ne 1 ] || [ ! -f "$output_path" ]; then
  echo "Timed out waiting for WebRecap.app output: $output_path" >&2
  exit 1
fi

if [ "$cleanup_tmp" -eq 1 ]; then
  cat "$output_path"
else
  echo "Wrote output to $output_path"
fi

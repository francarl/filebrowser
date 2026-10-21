#!/usr/bin/env bash
# Generate thumbnails for every video in a folder via filebrowser add-thumbnail.
set -euo pipefail

usage() {
  echo "Usage: $(basename "$0") --cacheDir <dir> [folder]" >&2
  echo "  --cacheDir  file cache directory (required)" >&2
  echo "  folder      folder to scan (default: current directory)" >&2
  exit 1
}

export PATH="$HOME/git/filebrowser:$PATH"

cacheDir=""
folder="."

while [[ $# -gt 0 ]]; do
  case $1 in
    --cacheDir)
      [[ $# -ge 2 ]] || usage
      cacheDir=$2
      shift 2
      ;;
    --cacheDir=*)
      cacheDir=${1#--cacheDir=}
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      folder=$1
      shift
      ;;
  esac
done

if [[ -z $cacheDir ]]; then
  echo "--cacheDir is required" >&2
  usage
fi

command -v filebrowser >/dev/null || {
  echo "filebrowser not found in PATH" >&2
  exit 1
}

if [[ ! -d $folder ]]; then
  echo "Folder not found: $folder" >&2
  exit 1
fi

shopt -s nullglob nocaseglob
found=0
for f in "$folder"/*.mp4 "$folder"/*.mkv; do
  found=1
  filebrowser add-thumbnail "$f" --cacheDir "$cacheDir"
done
shopt -u nullglob nocaseglob

if [[ $found -eq 0 ]]; then
  echo "No .mp4/.mkv files found in $folder" >&2
fi

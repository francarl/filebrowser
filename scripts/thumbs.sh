#!/usr/bin/env bash
# Generate <base>.thumbs.vtt + <base>.thumbs.jpg for a video, compatible with
# the videojs-vtt-thumbnails plugin used by filebrowser's preview player.
set -euo pipefail

usage() {
  echo "Usage: $(basename "$0") <file.mp4|file.mkv> [interval_seconds]" >&2
  exit 1
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
fi

file=$1
interval=${2:-5}

if [[ ! -f $file ]]; then
  echo "File not found: $file" >&2
  exit 1
fi
case $file in
  *.mp4 | *.mkv) ;;
  *)
    echo "Unsupported format (only .mp4/.mkv): $file" >&2
    exit 1
    ;;
esac
command -v ffmpeg >/dev/null || {
  echo "ffmpeg not found in PATH" >&2
  exit 1
}

W=120 # thumbnail width
H=68 # thumbnail height (16:9)
COLS=10 # sprite columns

duration=$(ffprobe -v error -show_entries format=duration \
  -of default=noprint_wrappers=1:nokey=1 "$file")
count=$(awk -v d="$duration" -v i="$interval" 'BEGIN { c = int(d / i) + 1; print (c < 1 ? 1 : c) }')
rows=$(( (count + COLS - 1) / COLS ))

base=${file%.*}
sprite="$base.thumbs.jpg"
vtt="$base.thumbs.vtt"

# One pass: sample one frame per interval, scale, tile into a single sprite.
ffmpeg -hide_banner -loglevel error -y -i "$file" \
  -vf "fps=1/$interval,scale=$W:$H,tile=${COLS}x${rows}" \
  -frames:v 1 -q:v 3 "$sprite"

{
  echo "WEBVTT"
  echo
  awk -v d="$duration" -v i="$interval" -v w="$W" -v h="$H" -v c="$COLS" 'BEGIN {
    for (k = 0; k * i < d; k++) {
      s = k * i
      e = s + i
      if (e > d) e = d
      printf "%02d:%02d:%06.3f --> %02d:%02d:%06.3f\n", \
        int(s / 3600), int((s % 3600) / 60), s % 60, \
        int(e / 3600), int((e % 3600) / 60), e % 60
      printf "#xywh=%d,%d,%d,%d\n", (k % c) * w, int(k / c) * h, w, h
      printf "\n"
    }
  }'
} >"$vtt"

echo "$sprite"
echo "$vtt"

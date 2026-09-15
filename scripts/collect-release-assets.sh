#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

out="${1:-release-assets}"
rm -rf "$out"
mkdir -p "$out"

mapfile -t apks < <(find fyne-cross/dist -type f -name '*.apk' | sort)
if [[ "${#apks[@]}" -ne 1 ]]; then
	echo "expected exactly one APK under fyne-cross/dist, found ${#apks[@]}" >&2
	find fyne-cross/dist -type f | sort >&2 || true
	exit 1
fi
cp "${apks[0]}" "$out/Stundenzettel-Generator.apk"

mapfile -t zips < <(find fyne-cross/dist -type f -name '*.zip' | grep -E '/windows' | sort)
mapfile -t exes < <(find fyne-cross/dist -type f -name '*.exe' | grep -E '/windows' | sort)

if [[ "${#zips[@]}" -ge 1 ]]; then
	cp "${zips[0]}" "$out/Stundenzettel-Generator-windows.zip"
elif [[ "${#exes[@]}" -ge 1 ]]; then
	win="${exes[0]}"
	(cd "$(dirname "$win")" && zip -j "$root/$out/Stundenzettel-Generator-windows.zip" "$(basename "$win")")
else
	echo "expected a Windows zip or exe under fyne-cross/dist" >&2
	find fyne-cross/dist -type f | sort >&2 || true
	exit 1
fi

ls -la "$out"

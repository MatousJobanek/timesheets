#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "usage: $0 X.Y.Z" >&2
	exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
	echo "working tree is not clean; commit or stash first" >&2
	exit 1
fi

if git rev-parse "v${version}" >/dev/null 2>&1; then
	echo "tag v${version} already exists" >&2
	exit 1
fi

build="$(python3 - "$version" <<'PY'
import re
import sys
from pathlib import Path

version = sys.argv[1]
path = Path("FyneApp.toml")
text = path.read_text()
match = re.search(r"^Build = (\d+)", text, re.M)
if not match:
    sys.exit("could not read Build from FyneApp.toml")
build = int(match.group(1)) + 1
text, n = re.subn(r'^Version = ".*"', f'Version = "{version}"', text, count=1, flags=re.M)
if n == 0:
    sys.exit(f'could not set Version to "{version}" in FyneApp.toml')
text, n = re.subn(r"^Build = \d+", f"Build = {build}", text, count=1, flags=re.M)
if n == 0:
    sys.exit(f"could not set Build to {build} in FyneApp.toml")
path.write_text(text)
print(build)
PY
)"

git add FyneApp.toml
git commit -m "Release v${version}"
git tag "v${version}"

echo "Committed Version ${version}, Build ${build}, tagged v${version}."
echo "Smoke-test on Linux, then: git push && git push --tags"

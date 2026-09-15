#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

tag="${1:-${GITHUB_REF_NAME:-}}"
if [[ -z "$tag" ]]; then
	echo "usage: $0 vX.Y.Z" >&2
	exit 1
fi
tag="${tag#v}"

version="$(sed -n 's/^Version = "\(.*\)"/\1/p' FyneApp.toml | head -n1)"
if [[ -z "$version" ]]; then
	echo "could not read Version from FyneApp.toml" >&2
	exit 1
fi
if [[ "$tag" != "$version" ]]; then
	echo "git tag v${tag} does not match FyneApp.toml Version ${version}" >&2
	exit 1
fi

echo "tag v${version} matches FyneApp.toml"

#!/bin/sh
set -eu
file="${1:-CHANGELOG.md}"
version=$(sed -n 's/^## \[\([0-9][^]]*\)\].*/\1/p' "$file" | head -n 1)
if [ -z "$version" ]; then
    echo "version.sh: no released version in $file" >&2
    exit 1
fi
printf '%s\n' "$version"

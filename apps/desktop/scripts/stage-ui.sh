#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
(cd ../client && bun install --frozen-lockfile && bun run build)
rm -rf dist
cp -R ../client/build dist
find dist -name '*.map' -delete
cp ui/popup.css ui/popup.js dist/
cp ui/index.html dist/popup.html

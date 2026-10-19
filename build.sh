#!/usr/bin/env bash
set -euo pipefail

BINARY=sqllings
PKG=./cmd/sqllings

# 모두 임시 폴더에 만든 뒤 마지막에 옮긴다.
# 커밋된 바이너리를 중간에 덮어쓰면 다음 빌드에 vcs.modified=true가 찍힌다.
OUT=$(mktemp -d)
trap 'rm -rf "${OUT}"' EXIT

echo "[1/4] mac arm64"
GOOS=darwin GOARCH=arm64 go build -trimpath -o "${OUT}/${BINARY}-darwin-arm64" "${PKG}"
echo "[2/4] mac amd64"
GOOS=darwin GOARCH=amd64 go build -trimpath -o "${OUT}/${BINARY}-darwin-amd64" "${PKG}"
echo "[3/4] mac 유니버설 (lipo)"
lipo -create -output "${OUT}/${BINARY}" "${OUT}/${BINARY}-darwin-arm64" "${OUT}/${BINARY}-darwin-amd64"
chmod +x "${OUT}/${BINARY}"
echo "[4/4] Windows amd64"
GOOS=windows GOARCH=amd64 go build -trimpath -o "${OUT}/${BINARY}.exe" "${PKG}"

mv "${OUT}/${BINARY}" "${OUT}/${BINARY}.exe" .

ls -lh "${BINARY}" "${BINARY}.exe"
file "${BINARY}"

#!/usr/bin/env bash
# Construye salida/InstalarMintDocente.exe (Windows, sin consola) desde Linux.
# Go se descarga a cache/go/ con versión y sha256 fijados.
set -euo pipefail

GO_VERSION=1.27.3
GO_TGZ_SHA256=65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40
GO_WINRES_VERSION=v0.3.3

AQUI=$(cd "$(dirname "$0")" && pwd)
RAIZ=$(dirname "$AQUI")
CACHE="$RAIZ/cache"
SALIDA="$RAIZ/salida"

if [ ! -x "$CACHE/go/bin/go" ] || [ "$("$CACHE/go/bin/go" env GOVERSION)" != "go$GO_VERSION" ]; then
  echo "Descargando Go $GO_VERSION"
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  curl -fsSL --retry 3 -o "$tmp/go.tgz" "https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz"
  echo "$GO_TGZ_SHA256  $tmp/go.tgz" | sha256sum -c -
  rm -rf "$CACHE/go"
  mkdir -p "$CACHE"
  tar -C "$CACHE" -xzf "$tmp/go.tgz"
fi
export PATH="$CACHE/go/bin:$PATH"
export GOPATH="$CACHE/gopath" GOFLAGS=-mod=mod GOTOOLCHAIN=local

cd "$AQUI"
go test ./...

# Recursos de Windows (manifiesto asInvoker, icono, nombre de producto)
go run "github.com/tc-hib/go-winres@$GO_WINRES_VERSION" make \
  --in winres/winres.json --out cmd/instalar/rsrc

mkdir -p "$SALIDA"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -ldflags "-H=windowsgui -s -w" \
  -o "$SALIDA/InstalarMintDocente.exe" ./cmd/instalar
echo "Listo: $SALIDA/InstalarMintDocente.exe"

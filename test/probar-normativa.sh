#!/bin/bash
# Prueba en el anfitrión (con red) el bloque «Normativa de Madrid» de provision/20-opencode.sh
# sin VM: venv y directorios temporales. Necesita uv o python3-venv.
#   test/probar-normativa.sh            -> debe terminar bien y mostrar los recuentos
#   test/probar-normativa.sh mal-sha    -> con un sha256 incorrecto el bloque debe fallar
set -euo pipefail
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
if command -v uv >/dev/null; then uv venv --seed -q "$T/v"; else python3 -m venv "$T/v"; fi
# Solo el bloque, sin las partes que exigen ser root
sed -n '/^log "Normativa de Madrid/,$p' "$RAIZ/provision/20-opencode.sh" > "$T/bloque.sh"
export NORMATIVA_DIR=$T/vistas NORMATIVA_BIN=$T/bin/normativa-madrid NORMATIVA_PY=$T/v/bin/python
[ "${1:-}" = mal-sha ] && export NORMATIVA_WHEEL_SHA256=$(printf 0%.0s {1..64})
bash -c 'set -euo pipefail; log() { echo "==> $*"; }; . "$1"' _ "$T/bloque.sh"
fichas=$("$NORMATIVA_BIN" catalogo | "$NORMATIVA_PY" -I -c 'import json,sys; print(len(json.load(sys.stdin)["fichas"]))')
vistas=$(ls "$NORMATIVA_DIR"/*.md | wc -l)
echo "fichas del catálogo: $fichas; vistas generadas: $vistas"
test "$fichas" -eq "$vistas" && test "$vistas" -gt 0

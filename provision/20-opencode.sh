#!/bin/bash
# OpenCode Desktop v2 (versión fijada), configuración del sistema, instrucciones y skills.
. /var/tmp/mint-docente/lib.sh || exit 1

OPENCODE_VERSION=2.2.5
OPENCODE_DEB_SHA256=c827e714220de744c829fdea55540ed814b1372061e7a905fde0978f9f819de8
GH_VERSION=2.103.3
GH_DEB_SHA256=e3bafef96328d1a67de3a968bdfe1cc2ffd58cf6aba8b5e201356897ed06bb3d

log "OpenCode Desktop $OPENCODE_VERSION"
tmp=$(mktemp -d)
curl -fsSL --retry 3 -o "$tmp/opencode.deb" \
  "https://github.com/anomalyco/opencode/releases/download/v$OPENCODE_VERSION/opencode-desktop-linux-amd64.deb"
echo "$OPENCODE_DEB_SHA256  $tmp/opencode.deb" | sha256sum -c -
apt_get install "$tmp/opencode.deb"
rm -rf "$tmp"
# CLI incluida en el paquete (la usa el asistente de primer arranque)
ln -sf /opt/OpenCode/resources/opencode-cli /usr/local/bin/opencode

log "GitHub CLI $GH_VERSION (para compartir con compañeros; la sesión es opcional)"
tmp=$(mktemp -d)
curl -fsSL --retry 3 -o "$tmp/gh.deb" \
  "https://github.com/cli/cli/releases/download/v$GH_VERSION/gh_${GH_VERSION}_linux_amd64.deb"
echo "$GH_DEB_SHA256  $tmp/gh.deb" | sha256sum -c -
apt_get install "$tmp/gh.deb"
rm -rf "$tmp"

log "Orden compartir (espacios compartidos entre docentes)"
install -D -m755 "$FILES/compartir/compartir" /usr/local/bin/compartir

log "Configuración del sistema (/etc/opencode) — prioridad máxima, sin credenciales"
install -D -m644 "$FILES/opencode/opencode.jsonc" /etc/opencode/opencode.jsonc

log "Instrucciones del agente y skills"
install -D -m644 "$FILES/opencode/AGENTS.md" "$SHARE/AGENTS.md"
rm -rf "$SHARE/skills"
cp -r "$FILES/opencode/skills" "$SHARE/skills"
chmod -R a+rX "$SHARE"

log "Normativa de Madrid (normativa-madrid 0.1.0)"
# Las variables permiten probar el bloque en un directorio temporal sin ser root.
NORMATIVA_DIR=${NORMATIVA_DIR:-$SHARE/normativa}
NORMATIVA_BIN=${NORMATIVA_BIN:-/usr/local/bin/normativa-madrid}
NORMATIVA_PY=${NORMATIVA_PY:-/opt/docente-py/bin/python}
NORMATIVA_WHEEL_URL=${NORMATIVA_WHEEL_URL:-https://github.com/balejosg/normativa-madrid/releases/download/v0.1.0/normativa_madrid-0.1.0-py3-none-any.whl}
NORMATIVA_WHEEL_SHA256=${NORMATIVA_WHEEL_SHA256:-09185b3b1920e8f85f3795f02c4525f06ba8e1aa299c72a3cf96d742854a4609}
tmp=$(mktemp -d)
wheel="$tmp/${NORMATIVA_WHEEL_URL##*/}"   # pip exige el nombre original del wheel
curl -fsSL --retry 3 -o "$wheel" "$NORMATIVA_WHEEL_URL"
echo "$NORMATIVA_WHEEL_SHA256  $wheel" | sha256sum -c -
"$NORMATIVA_PY" -m pip install --no-cache-dir "$wheel"
rm -rf "$tmp"
install -d "$(dirname "$NORMATIVA_BIN")"
ln -sf "$(dirname "$NORMATIVA_PY")/normativa-madrid" "$NORMATIVA_BIN"
# Una vista legible por ficha del catálogo (todas las etapas), para abrirla directamente.
# Nombre: «<código> - <nombre>.md» (hay nombres repetidos, el código los distingue).
rm -rf "$NORMATIVA_DIR"
install -d "$NORMATIVA_DIR"
"$NORMATIVA_BIN" catalogo | "$NORMATIVA_PY" -I -c '
import json, sys
for f in json.load(sys.stdin)["fichas"]:
    print(f["codigo"] + "\t" + f["nombre"])
' > "$NORMATIVA_DIR/.fichas"
while IFS=$'\t' read -r codigo nombre; do
  nombre=${nombre//\//-}; nombre=${nombre//:/ -}
  "$NORMATIVA_BIN" ficha "$codigo" --formato md > "$NORMATIVA_DIR/$codigo - $nombre.md"
done < "$NORMATIVA_DIR/.fichas"
rm -f "$NORMATIVA_DIR/.fichas"
chmod -R a+rX "$NORMATIVA_DIR"

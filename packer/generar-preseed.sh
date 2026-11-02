#!/bin/bash
# Genera packer/http/preseed.cfg a partir de packer/fuentes/preseed.cfg.in,
# incrustando packer/fuentes/postinstalacion.sh (gzip + base64, una línea).
set -euo pipefail
cd "$(dirname "$0")"
b64=$(grep -v '^# FUENTE\|^# Si lo cambias' fuentes/postinstalacion.sh | gzip -9n | base64 -w0)
mkdir -p http
sed "s|@POSTINSTALACION_B64@|$b64|" fuentes/preseed.cfg.in > http/preseed.cfg
# Comprobación: lo incrustado debe reproducir el script exactamente.
grep -o 'echo [A-Za-z0-9+/=]* | base64' http/preseed.cfg | cut -d' ' -f2 | base64 -d | gunzip \
  | diff -q - <(grep -v '^# FUENTE\|^# Si lo cambias' fuentes/postinstalacion.sh) >/dev/null
echo "preseed generado ($(wc -c < http/preseed.cfg) bytes, script ${#b64} caracteres base64)"

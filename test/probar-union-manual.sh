#!/bin/bash
# Prueba las órdenes del «Método manual» del README tal cual están escritas:
# 12 partes (part10 y part11 ponen a prueba el orden numérico), un
# SHA256SUMS como el de las releases, y una parte dañada que debe detectarse.
# El bloque de PowerShell solo se prueba si hay `pwsh`.
#   test/probar-union-manual.sh   -> debe terminar con «OK»
set -euo pipefail
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
fallo() { echo "FALLO: $1" >&2; exit 1; }

bloque() { # contenido del primer bloque ```$1 del README
  awk -v l="$1" '$0 == "```" l { d = 1; next } d && /^```/ { exit } d' "$RAIZ/README.md"
}
bloque bash > "$T/unir.sh"
bloque powershell > "$T/unir.ps1"
[ -s "$T/unir.sh" ] && [ -s "$T/unir.ps1" ] || fallo "no encuentro los bloques en el README"
! grep -q 'part\*.*>' "$T/unir.sh" || fallo "el README une con un patrón part*"

OVA=MintDocente-AAAAMMDD.ova
preparar() { # $1: carpeta; deja 12 partes distintas y su SHA256SUMS
  mkdir -p "$1"
  for i in $(seq 0 11); do head -c $((1000 + i)) /dev/urandom > "$1/$OVA.part$i"; done
  (cd "$1"
   for i in $(seq 0 11); do sha256sum "$OVA.part$i"; done > SHA256SUMS
   for i in $(seq 0 11); do cat "$OVA.part$i"; done | sha256sum | sed "s/-\$/$OVA/" >> SHA256SUMS)
}

# bash: bien y con una parte dañada
preparar "$T/sh-bien"
(cd "$T/sh-bien" && bash "$T/unir.sh") > "$T/sh-bien.out" 2>&1 || { cat "$T/sh-bien.out"; fallo "bash: la unión correcta no pasa la comprobación"; }
grep -q ': OK' "$T/sh-bien.out" || fallo "bash: no dice OK"
preparar "$T/sh-mal"
printf 'x' >> "$T/sh-mal/$OVA.part10"
! (cd "$T/sh-mal" && bash "$T/unir.sh") > "$T/sh-mal.out" 2>&1 || fallo "bash: no detecta la parte dañada"

if command -v pwsh >/dev/null; then
  preparar "$T/ps-bien"
  (cd "$T/ps-bien" && pwsh -NoProfile -File "$T/unir.ps1") > "$T/ps-bien.out" 2>&1
  grep -qx 'OVA correcta' "$T/ps-bien.out" || { cat "$T/ps-bien.out"; fallo "PowerShell: la unión correcta no pasa"; }
  cmp -s "$T/ps-bien/$OVA" <(cd "$T/ps-bien" && for i in $(seq 0 11); do cat "$OVA.part$i"; done) \
    || fallo "PowerShell: la OVA no sale en orden numérico"
  preparar "$T/ps-mal"
  printf 'x' >> "$T/ps-mal/$OVA.part10"
  (cd "$T/ps-mal" && pwsh -NoProfile -File "$T/unir.ps1") > "$T/ps-mal.out" 2>&1
  grep -q 'NO coincide' "$T/ps-mal.out" || fallo "PowerShell: no detecta la parte dañada"
else
  echo "(sin pwsh: no se prueba el bloque de Windows)"
fi
echo OK

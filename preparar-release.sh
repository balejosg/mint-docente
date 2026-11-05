#!/usr/bin/env bash
# Prepara los archivos de una release en salida/release-<etiqueta>/:
#   - la OVA más reciente de salida/ (MintDocente-AAAAMMDD.ova) en partes
#     .ova.part0, .part1… de menos de 2 GB (límite de GitHub);
#   - SHA256SUMS: una línea por parte y una para la .ova completa (es lo que
#     lee el instalador);
#   - InstalarMintDocente.exe (lo genera instalador/construir.sh).
# No sube nada. Para publicarla, ver «Preparar una release» en README.md.
#
# Uso: ./preparar-release.sh <etiqueta> [ruta de la .ova]
#   TAM_PARTE  bytes por parte (por defecto 1992294400, como en v1.1)
#   SALIDA     carpeta de entrada y salida (por defecto ./salida)
set -euo pipefail

TAM_PARTE=${TAM_PARTE:-1992294400}
AVISO_BYTES=$((38 * 1024 * 1024 * 1024 / 10))   # 3,8 GB

AQUI=$(cd "$(dirname "$0")" && pwd)
SALIDA=${SALIDA:-$AQUI/salida}

fallo() { echo "preparar-release: $*" >&2; exit 1; }

if [ $# -lt 1 ] || [ $# -gt 2 ]; then fallo "uso: $0 <etiqueta> [ruta de la .ova]"; fi
ETIQUETA=$1
[[ "$ETIQUETA" =~ ^[A-Za-z0-9._-]+$ ]] || fallo "etiqueta no válida: $ETIQUETA"
[[ "$TAM_PARTE" =~ ^[1-9][0-9]*$ ]] || fallo "TAM_PARTE no válido: $TAM_PARTE"

if [ $# -eq 2 ]; then
  OVA=$2
else
  OVA=$(find "$SALIDA" -maxdepth 1 -name 'MintDocente-*.ova' | sort | tail -n 1)
  [ -n "$OVA" ] || fallo "no hay ninguna MintDocente-*.ova en $SALIDA (ejecuta ./build.sh)"
fi
[ -f "$OVA" ] || fallo "no existe $OVA"
NOMBRE=$(basename "$OVA")
[[ "$NOMBRE" == *.ova ]] || fallo "$NOMBRE no es una .ova"

EXE=$SALIDA/InstalarMintDocente.exe
[ -f "$EXE" ] || fallo "no existe $EXE (ejecuta instalador/construir.sh)"

DEST=$SALIDA/release-$ETIQUETA
[ ! -e "$DEST" ] || fallo "$DEST ya existe; bórrala o usa otra etiqueta"

# Si build.sh dejó la suma de la OVA, se comprueba antes de trocearla.
if [ -f "$OVA.sha256" ]; then
  esperada=$(awk '{print $1; exit}' "$OVA.sha256")
  echo "$esperada  $OVA" | sha256sum -c --status - 2>/dev/null || fallo "$NOMBRE no coincide con $NOMBRE.sha256"
fi

TAM=$(stat -c %s "$OVA")
[ "$TAM" -gt 0 ] || fallo "$NOMBRE está vacía"
if [ "$TAM" -gt "$AVISO_BYTES" ]; then
  echo "AVISO: $NOMBRE ocupa $((TAM / 1000000)) MB, más de 3,8 GB. Revísalo antes de publicar." >&2
fi

tmp=$DEST.tmp
rm -rf "$tmp"
mkdir -p "$tmp"
trap 'rm -rf "$tmp"' EXIT

# split numera con ceros (part0000); se renombran a part0, part1…, part10.
split -b "$TAM_PARTE" -d -a 4 "$OVA" "$tmp/$NOMBRE.part"
n=0
while [ -e "$tmp/$NOMBRE.part$(printf %04d "$n")" ]; do
  mv "$tmp/$NOMBRE.part$(printf %04d "$n")" "$tmp/$NOMBRE.part$n.nuevo"
  n=$((n + 1))
done
for ((i = 0; i < n; i++)); do
  mv "$tmp/$NOMBRE.part$i.nuevo" "$tmp/$NOMBRE.part$i"
done

(
  cd "$tmp"
  for ((i = 0; i < n; i++)); do sha256sum "$NOMBRE.part$i"; done
) > "$tmp/SHA256SUMS"
(cd "$(dirname "$OVA")" && sha256sum "$NOMBRE") >> "$tmp/SHA256SUMS"
(cd "$tmp" && sha256sum -c --quiet --strict SHA256SUMS --ignore-missing) \
  || fallo "las partes no coinciden con SHA256SUMS"

cp "$EXE" "$tmp/InstalarMintDocente.exe"
mv "$tmp" "$DEST"
trap - EXIT

echo "Release $ETIQUETA lista en $DEST ($n partes):"
ls -l "$DEST"

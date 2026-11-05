#!/bin/bash
# Prueba en el anfitrión de mint-docente-desconectar, sin VM ni zenity/gh reales:
# HOME temporal y programas falsos en el PATH.
#   test/probar-desconectar.sh   -> debe terminar con «OK»
set -euo pipefail
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
S="$RAIZ/files/escritorio/mint-docente-desconectar"
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
mkdir -p "$T/bin" "$T/home/.local/share/opencode"
export HOME="$T/home" LLAMADAS="$T/llamadas" PATH="$T/bin:$PATH"
cat > "$T/bin/gh" <<'F'
#!/bin/bash
echo "$*" >> "$LLAMADAS"
F
cat > "$T/bin/zenity" <<'F'
#!/bin/bash
echo "zenity $*" >> "$LLAMADAS"
case "$*" in *--question*) exit "${ZENITY_RESPUESTA:-1}";; esac
F
chmod +x "$T/bin/gh" "$T/bin/zenity"
fallo() { echo "FALLO: $1" >&2; exit 1; }

# 1. El docente dice que no: no se toca nada
echo '{}' > "$HOME/.local/share/opencode/auth.json"; : > "$LLAMADAS"
ZENITY_RESPUESTA=1 bash "$S" || true
[ -e "$HOME/.local/share/opencode/auth.json" ] || fallo "borró auth.json sin permiso"
! grep -q 'auth logout' "$LLAMADAS" || fallo "cerró GitHub sin permiso"

# 2. El docente acepta (zenity responde sí)
: > "$LLAMADAS"
ZENITY_RESPUESTA=0 bash "$S"
[ ! -e "$HOME/.local/share/opencode/auth.json" ] || fallo "auth.json sigue ahí (pregunta)"
grep -q 'auth logout -h github.com' "$LLAMADAS" || fallo "no cerró GitHub (pregunta)"

# 3. --si se salta la pregunta
echo '{}' > "$HOME/.local/share/opencode/auth.json"; : > "$LLAMADAS"
bash "$S" --si
[ ! -e "$HOME/.local/share/opencode/auth.json" ] || fallo "auth.json sigue ahí (--si)"
grep -q 'auth logout -h github.com' "$LLAMADAS" || fallo "no cerró GitHub (--si)"
! grep -q -- '--question' "$LLAMADAS" || fallo "--si preguntó"
grep -q -- '--info' "$LLAMADAS" || fallo "sin aviso final"
echo OK

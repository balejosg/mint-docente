#!/bin/bash
# Prueba en el anfitrión de mint-docente-desconectar, sin VM ni zenity/gh reales:
# HOME temporal y programas falsos en el PATH.
#   test/probar-desconectar.sh   -> debe terminar con «OK»
#
# El gh falso guarda la sesión en ~/.config/gh/hosts.yml, como el real con
# --insecure-storage. Variables para simular fallos:
#   GH_STATUS_FALLA=1   «gh auth status» falla (sin red, token caducado…)
#   GH_LOGOUT_FALLA=1   «gh auth logout» falla sin tocar nada
#   GH_LLAVERO=1        el token sigue en otro sitio: «gh auth token» lo da siempre
set -euo pipefail
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
S="$RAIZ/files/escritorio/mint-docente-desconectar"
T=$(mktemp -d); trap 'chmod -R u+w "$T"; rm -rf "$T"' EXIT
mkdir -p "$T/bin" "$T/home/.local/share/opencode"
export HOME="$T/home" LLAMADAS="$T/llamadas" PATH="$T/bin:$PATH"
unset GH_CONFIG_DIR XDG_CONFIG_HOME GH_TOKEN GITHUB_TOKEN
HOSTS="$HOME/.config/gh/hosts.yml"
cat > "$T/bin/gh" <<'F'
#!/bin/bash
echo "$*" >> "$LLAMADAS"
h="$HOME/.config/gh/hosts.yml"
case "$*" in
  "auth status"*) [ -z "${GH_STATUS_FALLA:-}" ] && grep -q '^github.com:' "$h" 2>/dev/null ;;
  "auth logout"*)
    [ -z "${GH_LOGOUT_FALLA:-}" ] || exit 1
    grep -q '^github.com:' "$h" 2>/dev/null || exit 1
    rm -f "$h" ;;
  "auth token"*)
    if [ -n "${GH_LLAVERO:-}" ] || grep -q '^github.com:' "$h" 2>/dev/null; then echo gho_secreto; else exit 1; fi ;;
esac
F
cat > "$T/bin/zenity" <<'F'
#!/bin/bash
echo "zenity $*" >> "$LLAMADAS"
case "$*" in *--question*) exit "${ZENITY_RESPUESTA:-1}";; esac
F
chmod +x "$T/bin/gh" "$T/bin/zenity"
fallo() { echo "FALLO: $1" >&2; exit 1; }

sesion() { # deja una sesión de GitHub (y opcionalmente otro servidor) y la de IA
  mkdir -p "$HOME/.config/gh"
  printf 'github.com:\n    users:\n        profe:\n            oauth_token: gho_secreto\n    git_protocol: https\n    oauth_token: gho_secreto\n    user: profe\n' > "$HOSTS"
  [ "${1:-}" != "otro" ] || printf 'ghe.ejemplo.es:\n    oauth_token: gho_otro\n    user: profe2\n' >> "$HOSTS"
  echo '{}' > "$HOME/.local/share/opencode/auth.json"
  : > "$LLAMADAS"
}
exito()   { grep -q -- '--info.*desconectadas' "$LLAMADAS"; }
aviso_error() { grep -q -- '--error' "$LLAMADAS"; }

# 1. El docente dice que no: no se toca nada
sesion
ZENITY_RESPUESTA=1 bash "$S" || true
[ -e "$HOME/.local/share/opencode/auth.json" ] || fallo "borró auth.json sin permiso"
[ -e "$HOSTS" ] || fallo "borró la sesión de GitHub sin permiso"
! grep -q 'auth logout' "$LLAMADAS" || fallo "cerró GitHub sin permiso"

# 2. El docente acepta (zenity responde sí)
sesion
ZENITY_RESPUESTA=0 bash "$S"
[ ! -e "$HOME/.local/share/opencode/auth.json" ] || fallo "auth.json sigue ahí (pregunta)"
grep -q 'auth logout -h github.com' "$LLAMADAS" || fallo "no cerró GitHub (pregunta)"
[ ! -e "$HOSTS" ] || fallo "el token sigue en hosts.yml (pregunta)"
exito || fallo "sin aviso de éxito (pregunta)"

# 3. --si se salta la pregunta
sesion
bash "$S" --si
[ ! -e "$HOME/.local/share/opencode/auth.json" ] || fallo "auth.json sigue ahí (--si)"
grep -q 'auth logout -h github.com' "$LLAMADAS" || fallo "no cerró GitHub (--si)"
! grep -q -- '--question' "$LLAMADAS" || fallo "--si preguntó"
exito || fallo "sin aviso final (--si)"

# 4. Sin red: «gh auth status» falla, pero hay token. Hay que cerrar igual.
sesion
GH_STATUS_FALLA=1 bash "$S" --si
grep -q 'auth logout -h github.com' "$LLAMADAS" || fallo "sin red no intentó cerrar GitHub"
[ ! -e "$HOSTS" ] || fallo "sin red el token sigue en hosts.yml"
exito || fallo "sin red no avisó del éxito"

# 5. «gh auth logout» falla: se borra la entrada de github.com a mano
sesion
GH_STATUS_FALLA=1 GH_LOGOUT_FALLA=1 bash "$S" --si
[ ! -e "$HOSTS" ] || fallo "con el logout fallido el token sigue en hosts.yml"
exito || fallo "con el logout fallido no avisó del éxito"

# 6. Igual, pero hay otro servidor: solo se quita github.com
sesion otro
GH_LOGOUT_FALLA=1 bash "$S" --si
[ -e "$HOSTS" ] || fallo "borró la sesión de otro servidor"
! grep -q 'github.com' "$HOSTS" || fallo "github.com sigue en hosts.yml (con otro servidor)"
! grep -q 'gho_secreto' "$HOSTS" || fallo "el token de github.com sigue en hosts.yml"
grep -q 'gho_otro' "$HOSTS" || fallo "perdió el token del otro servidor"
exito || fallo "con otro servidor no avisó del éxito"

# 7. El archivo del token no se puede borrar: nada de «✔», sino un aviso claro
sesion
chmod a-w "$HOME/.config/gh"
r=0; GH_LOGOUT_FALLA=1 bash "$S" --si || r=$?
chmod u+w "$HOME/.config/gh"
[ "$r" -ne 0 ] || fallo "salió con 0 sin poder borrar el token"
! exito || fallo "dijo «desconectadas» con el token aún en el equipo"
aviso_error || fallo "no avisó de que GitHub sigue conectado"
grep -q -- '--error.*GitHub' "$LLAMADAS" || fallo "el aviso no menciona GitHub"
! grep -q -- '--error.*cuenta de IA' "$LLAMADAS" || fallo "avisó de la IA sin motivo"

# 8. gh sigue dando un token (guardado en otro sitio): tampoco hay «✔»
sesion
r=0; GH_LLAVERO=1 bash "$S" --si || r=$?
[ "$r" -ne 0 ] || fallo "salió con 0 con gh dando aún un token"
! exito || fallo "dijo «desconectadas» con gh dando aún un token"
grep -q -- '--error.*GitHub' "$LLAMADAS" || fallo "no avisó de GitHub (token en otro sitio)"

# 9. auth.json no se puede borrar: avisa de la cuenta de IA
sesion
chmod a-w "$HOME/.local/share/opencode"
r=0; bash "$S" --si || r=$?
chmod u+w "$HOME/.local/share/opencode"
[ "$r" -ne 0 ] || fallo "salió con 0 sin poder borrar auth.json"
! exito || fallo "dijo «desconectadas» con auth.json aún en el equipo"
grep -q -- '--error.*cuenta de IA' "$LLAMADAS" || fallo "no avisó de la cuenta de IA"
! grep -q -- '--error.*GitHub' "$LLAMADAS" || fallo "avisó de GitHub sin motivo"

# 10. Sin ninguna sesión abierta: éxito, sin errores
rm -rf "$HOME/.config/gh" "$HOME/.local/share/opencode/auth.json"; : > "$LLAMADAS"
bash "$S" --si
exito || fallo "sin sesiones no avisó del éxito"
echo OK

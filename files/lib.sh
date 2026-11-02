# Utilidades comunes para los scripts de aprovisionamiento.
# Se incluye con: . "$(dirname "$0")/lib.sh"  (o desde /var/tmp/mint-docente/lib.sh)
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=a

DOCENTE_USER=docente
DOCENTE_HOME=/home/$DOCENTE_USER
FILES=/var/tmp/mint-docente
SHARE=/usr/share/mint-docente

log() { printf '\n\033[1;34m==> %s\033[0m\n' "$*"; }

apt_get() {
  apt-get -y -o DPkg::Lock::Timeout=900 \
    -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold "$@"
}

# Ejecuta un comando como el usuario docente
como_docente() { sudo -u "$DOCENTE_USER" -H "$@"; }

# Instala un archivo propiedad de docente
instalar_docente() { # origen destino [modo]
  install -D -o "$DOCENTE_USER" -g "$DOCENTE_USER" -m "${3:-644}" "$1" "$2"
}

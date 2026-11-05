#!/bin/bash
# Imagen más ligera: un solo kernel retenido, sin compiladores, sin programas
# que el docente no usa, sin firmware y solo con los idiomas de España e inglés.
. /var/tmp/mint-docente/lib.sh || exit 1

KERNEL=$(uname -r)                 # p. ej. 7.0.0-42-generic (el más nuevo tras 00-base)
KVER=${KERNEL%-generic}            # 7.0.0-42
VERSION_KERNEL='[0-9]+\.[0-9]+\.[0-9]+-[0-9]+'

# Paquetes instalados de una lista de nombres o patrones (los demás se ignoran).
instalados() {
  { dpkg-query -W -f='${db:Status-Abbrev} ${Package}\n' "$@" 2>/dev/null || true; } | awk '$1 ~ /^.i/ {print $2}'
}

# Purga los paquetes dados solo si apt no se lleva nada más con ellos.
purgar_solo() {
  [ $# -gt 0 ] || return 0
  local fuera
  fuera=$(LANG=C apt-get -s purge "$@" 2>/dev/null | awk '/^Purg /{print $2}' | grep -vxF -f <(printf '%s\n' "$@") || true)
  if [ -n "$fuera" ]; then
    echo "Se quedan $* (se llevarían también: $(echo $fuera))"
  else
    apt_get purge "$@"
  fi
}

log "Guest Additions sin compilador: los módulos del kernel $KERNEL ya están instalados"
for m in vboxguest vboxsf; do
  ruta=$(modinfo -k "$KERNEL" -n "$m" 2>/dev/null || true)
  case "$ruta" in
    /lib/modules/"$KERNEL"/*|/usr/lib/modules/"$KERNEL"/*) echo "$m: $ruta" ;;
    *) echo "Falta el módulo $m para $KERNEL: no se pueden quitar los compiladores"; exit 1 ;;
  esac
done

log "Quitar los kernels que no están en marcha"
# Versión exacta: con 6.8.0-4 en marcha, 6.8.0-45 también es otro kernel.
version_en_nombre() { grep -oE "$VERSION_KERNEL" <<<"$1" | tail -n1; }
paquetes_kernel() { # paquetes linux-* instalados con una versión de kernel en el nombre
  dpkg-query -W -f='${db:Status-Abbrev} ${Package}\n' 'linux-*' | awk '$1 ~ /^.i/ {print $2}' \
    | { grep -E "^linux-.*$VERSION_KERNEL" || true; }
}
VIEJOS=$(paquetes_kernel | while read -r p; do
  if [ "$(version_en_nombre "$p")" != "$KVER" ]; then echo "$p"; fi
done)
[ -z "$VIEJOS" ] || apt_get purge $VIEJOS

log "Quitar compiladores, DKMS y las cabeceras del kernel en marcha"
apt_get purge $(instalados build-essential dkms "linux-headers-$KERNEL")
# Los compiladores están marcados como manuales en Mint: purgar build-essential no
# los quita. cpp se queda (lo necesita x11-xserver-utils).
purgar_solo $(instalados gcc g++ 'gcc-*' 'g++-*' 'libstdc++-*-dev' 'libgcc-*-dev' \
  | grep -E '^(gcc|g\+\+)(-[0-9]+)?(-x86-64-linux-gnu)?$|^lib(stdc\+\+|gcc)-[0-9]+-dev$')

log "Retener el kernel: sin compiladores, uno nuevo dejaría sin Guest Additions"
METAS=$(instalados linux-generic-hwe-24.04 linux-image-generic-hwe-24.04 linux-headers-generic-hwe-24.04 \
  linux-generic linux-image-generic linux-headers-generic)
[ -n "$METAS" ] || { echo "No hay ningún metapaquete del kernel que retener"; exit 1; }
# Manuales para que autoremove no se los lleve al quitar linux-generic-hwe-24.04.
apt-mark manual $METAS
apt-mark hold $METAS
# La retención solo cubre los metapaquetes: la ventana de kernels del Gestor de
# actualizaciones instala paquetes con versión (linux-image-X) directamente.
# Se bloquean todos (prioridad -1) salvo los del kernel en marcha, que van
# detrás con su nombre exacto: apt_preferences(5) pide los patrones primero
# para que los nombres exactos de después manden sobre ellos.
EN_MARCHA=$(paquetes_kernel | while read -r p; do
  if [ "$(version_en_nombre "$p")" = "$KVER" ]; then echo "$p"; fi
done | tr '\n' ' ')
[ -n "$EN_MARCHA" ] || { echo "No se encuentran los paquetes del kernel en marcha"; exit 1; }
cat > /etc/apt/preferences.d/90-mint-docente-kernel <<EOF
# Mint Docente: sin compiladores, un kernel nuevo dejaría sin Guest Additions.
Package: linux-image-[0-9]* linux-image-unsigned-[0-9]* linux-modules-[0-9]* linux-modules-extra-[0-9]* linux-headers-[0-9]* linux-main-modules-* linux-hwe-*
Pin: version *
Pin-Priority: -1

Package: $EN_MARCHA
Pin: version *
Pin-Priority: 500
EOF
cat /etc/apt/preferences.d/90-mint-docente-kernel

log "Sin actualizaciones automáticas"
systemctl disable --now mintupdate-automation-upgrade.timer 2>/dev/null || true
if [ -n "$(instalados unattended-upgrades)" ]; then
  systemctl disable --now unattended-upgrades.service 2>/dev/null || true
  printf 'APT::Periodic::Unattended-Upgrade "0";\n' > /etc/apt/apt.conf.d/99mint-docente-sin-auto
fi

log "Programas que el docente no usa"
purgar_solo $(instalados 'thunderbird*' 'hexchat*' 'transmission-*' hypnotix celluloid 'rhythmbox*' \
  warpinator timeshift)

log "Fuentes de otros alfabetos (Noto Sans de latín, griego y cirílico se queda)"
purgar_solo $(instalados 'fonts-noto-cjk*' fonts-noto-extra fonts-noto-ui-extra)

log "Temas de iconos: solo el del docente, sus temas de respaldo y hicolor"
# Valor en vigor de xfconf: el del usuario si lo ha cambiado; si no, el de Mint
# (mint-artwork) y, en último caso, el de Xfce. Sin sesión gráfica no hay xfconfd.
xsettings() { # propiedad → primer valor que es un tema instalado
  local f v
  for f in "$DOCENTE_HOME/.config/xfce4/xfconf/xfce-perchannel-xml/xsettings.xml" \
           /usr/share/mint-artwork/xfce/xfce4/xfconf/xfce-perchannel-xml/xsettings.xml \
           /etc/xdg/xdg-xfce/xfce4/xfconf/xfce-perchannel-xml/xsettings.xml \
           /etc/xdg/xfce4/xfconf/xfce-perchannel-xml/xsettings.xml; do
    v=$(sed -n "s/.*name=\"$1\" type=\"string\" value=\"\([^\"]*\)\".*/\1/p" "$f" 2>/dev/null | head -n1)
    if [ -n "$v" ] && [ -d "/usr/share/icons/$v" ]; then echo "$v"; return; fi
  done
}
ICONOS=$(xsettings IconThemeName)
CURSOR=$(xsettings CursorThemeName)
[ -n "$ICONOS" ] || { echo "No se encuentra el tema de iconos del docente"; exit 1; }
# El tema y su cadena de respaldo (Inherits), hicolor, el cursor y «default».
CONSERVAR="hicolor default${CURSOR:+ $CURSOR}"
pendientes=$ICONOS
while [ -n "$pendientes" ]; do
  set -- $pendientes; t=$1; shift; pendientes="$*"
  case " $CONSERVAR " in *" $t "*) continue ;; esac
  CONSERVAR="$CONSERVAR $t"
  for h in $(sed -n 's/^Inherits=//p' "/usr/share/icons/$t/index.theme" 2>/dev/null | tr ',' ' '); do
    if [ -d "/usr/share/icons/$h" ]; then pendientes="$pendientes $h"; fi
  done
done
echo "Se conservan: $CONSERVAR"
# Paquetes que ponen algún tema conservado: no se tocan.
NECESARIOS=$(for t in $CONSERVAR; do dpkg -S "/usr/share/icons/$t" 2>/dev/null | cut -d: -f1 | tr ',' '\n'; done \
  | tr -d ' ' | sort -u)
CANDIDATOS=$(for d in /usr/share/icons/*/; do dpkg -S "${d%/}" 2>/dev/null | cut -d: -f1 | tr ',' '\n'; done \
  | tr -d ' ' | sort -u | grep -vxF -f <(echo "$NECESARIOS") || true)
# Solo se purga un paquete de temas (sin programas ni lanzadores, p. ej. no onboard)
# y si al hacerlo no se lleva nada más que otros temas.
SOLO_TEMAS=""
for p in $CANDIDATOS; do
  # La lista se guarda antes de buscar: con grep -q y pipefail, dpkg -L podría
  # morir por SIGPIPE y un programa pasaría por tema.
  archivos=$(dpkg -L "$p" 2>/dev/null || true)
  if grep -E '^/usr/(s?bin|lib|libexec|games|share/applications)/' <<<"$archivos" >/dev/null; then
    echo "Se queda $p (no es solo un tema)"
  else
    SOLO_TEMAS="$SOLO_TEMAS $p"
  fi
done
# Lo que apt se lleve con cada uno tiene que ser también solo temas (nunca onboard).
TEMAS=""
for p in $SOLO_TEMAS; do
  fuera=$(LANG=C apt-get -s purge "$p" 2>/dev/null | awk '/^Purg /{print $2}' \
    | grep -vxF -f <(printf '%s\n' $SOLO_TEMAS) || true)
  if [ -z "$fuera" ]; then TEMAS="$TEMAS $p"; else echo "Se queda $p (lo necesitan: $(echo $fuera))"; fi
done
[ -z "$TEMAS" ] || apt_get purge $TEMAS
# La pantalla de inicio de sesión usa el mismo tema de iconos que el docente.
if [ -d /etc/lightdm/lightdm-gtk-greeter.conf.d ]; then
  printf '[greeter]\nicon-theme-name=%s\n' "$ICONOS" > /etc/lightdm/lightdm-gtk-greeter.conf.d/99_mint-docente.conf
fi

log "Sin firmware, documentación, manuales ni idiomas que no se usan (también en las actualizaciones)"
IDIOMAS="es ca gl eu en"
{
  echo "# Mint Docente: una VM no necesita firmware; solo idiomas de España e inglés."
  echo "path-exclude=/lib/firmware/*"
  echo "path-exclude=/usr/lib/firmware/*"
  echo "path-exclude=/usr/share/doc/*"
  echo "path-include=/usr/share/doc/*/copyright"
  echo "path-exclude=/usr/share/man/*"
  echo "path-exclude=/usr/share/locale/*"
  echo "path-include=/usr/share/locale/locale.alias"
  for i in $IDIOMAS; do echo "path-include=/usr/share/locale/$i*"; echo "path-include=/usr/share/man/$i*"; done
} > /etc/dpkg/dpkg.cfg.d/90-mint-docente
cat /etc/dpkg/dpkg.cfg.d/90-mint-docente
# Lo que ya está instalado, con las mismas reglas.
rm -rf /usr/lib/firmware/*
find /usr/share/doc -mindepth 1 ! -type d ! -name copyright -delete
find /usr/share/doc -mindepth 1 -type d -empty -delete
borrar_salvo_idiomas() { # directorio [nombre que también se conserva]
  local e n
  for e in "$1"/*; do
    n=${e##*/}
    [ "$n" = "${2:-}" ] && continue
    case "$n" in es*|ca*|gl*|eu*|en*) continue ;; esac
    rm -rf "$e"
  done
}
borrar_salvo_idiomas /usr/share/locale locale.alias
borrar_salvo_idiomas /usr/share/man

log "Fondos: solo el del escritorio y el de la pantalla de inicio de sesión"
FONDOS=$(
  for f in "$DOCENTE_HOME/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-desktop.xml" \
           /usr/share/mint-artwork/xfce/xfce4/xfconf/xfce-perchannel-xml/xfce4-desktop.xml \
           /etc/xdg/xdg-xfce/xfce4/xfconf/xfce-perchannel-xml/xfce4-desktop.xml \
           /etc/xdg/xfce4/xfconf/xfce-perchannel-xml/xfce4-desktop.xml; do
    if [ -f "$f" ]; then
      sed -n 's/.*name="\(last-image\|image-path\)" type="string" value="\([^"]*\)".*/\2/p' "$f"
    fi
  done
  # Puede no haber ningún .conf (el de 90-limpieza aún no existe): sin globs vacíos.
  for f in /etc/lightdm/*.conf /etc/lightdm/*.conf.d/*.conf; do
    if [ -f "$f" ]; then sed -n 's/^ *background *= *//p' "$f"; fi
  done
)
CONSERVAR_FONDOS=$(printf '%s\n' "$FONDOS" | while IFS= read -r f; do
  if [ -n "$f" ] && [ -e "$f" ]; then readlink -f "$f"; fi
done | sort -u)
if [ -z "$CONSERVAR_FONDOS" ]; then
  # Sin saber cuál es el fondo no se borra ninguno: mejor pesar más que un escritorio vacío.
  echo "No se encuentra el fondo del escritorio: se conservan todos los fondos"
else
  echo "Se conservan:"; printf '  %s\n' $CONSERVAR_FONDOS
  find /usr/share/backgrounds -type f -print0 | while IFS= read -r -d '' f; do
    if ! grep -qxF -- "$f" <<<"$CONSERVAR_FONDOS"; then rm -f "$f"; fi
  done
  # Enlaces que apuntaban a los fondos borrados (p. ej. en /usr/share/xfce4/backdrops).
  find /usr/share/backgrounds /usr/share/xfce4/backdrops -xtype l -delete 2>/dev/null || true
  find /usr/share/backgrounds -mindepth 1 -type d -empty -delete
fi

log "Restos de los paquetes quitados"
apt_get autoremove --purge
apt-get clean

log "Reinicio para comprobar que la máquina arranca con lo que queda"
(sleep 3; systemctl reboot) &
exit 0

#!/bin/bash
# Sistema base: Guest Additions, actualización, idioma español, escritorio sin bloqueos.
. /tmp/mint-docente/lib.sh

log "Desactivar el autologin mientras dura el build (se reactiva en 90-limpieza)"
sed -i -E 's/^(autologin-user|autologin-user-timeout)=.*/#&/' /etc/lightdm/lightdm.conf 2>/dev/null || true

log "Esperar a que terminen procesos apt de arranque"
systemctl stop unattended-upgrades.service mintupdate-automation-upgrade.timer 2>/dev/null || true

log "Dependencias de compilación para las Guest Additions"
apt_get update
apt_get install build-essential dkms perl "linux-headers-$(uname -r)"

log "Guest Additions de VirtualBox (versión del anfitrión que construye la VM)"
mkdir -p /mnt/vbga
mount -o loop,ro /tmp/VBoxGuestAdditions.iso /mnt/vbga
sh /mnt/vbga/VBoxLinuxAdditions.run --nox11 || true   # devuelve 2 si pide reinicio
umount /mnt/vbga && rm -f /tmp/VBoxGuestAdditions.iso
usermod -aG vboxsf "$DOCENTE_USER"
test -e /sbin/rcvboxadd || { echo "Guest Additions no instaladas"; exit 1; }

log "Actualización completa del sistema"
apt_get full-upgrade
apt_get install linux-headers-generic-hwe-24.04 2>/dev/null || apt_get install linux-headers-generic

log "Idioma español completo"
apt_get install $(check-language-support -l es) language-pack-es language-pack-gnome-es \
  hunspell-es hyphen-es mythes-es libreoffice-l10n-es libreoffice-help-es
locale-gen es_ES.UTF-8
update-locale LANG=es_ES.UTF-8 LANGUAGE=es_ES:es

log "Sin salvapantallas ni bloqueo de pantalla (lo gestiona el anfitrión)"
apt_get purge xfce4-screensaver light-locker 2>/dev/null || true

log "Sin ventana de bienvenida de Mint"
como_docente mkdir -p "$DOCENTE_HOME/.linuxmint/mintwelcome"
como_docente touch "$DOCENTE_HOME/.linuxmint/mintwelcome/norun.flag"

log "Reinicio para cargar kernel y Guest Additions"
(sleep 3; systemctl reboot) &
exit 0

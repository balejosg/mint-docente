#!/bin/bash
# Limpieza final: autologin, sin restos del build, sin credenciales, disco compacto.
. /var/tmp/mint-docente/lib.sh || exit 1

log "Inicio de sesión automático como docente"
install -D -m644 /dev/stdin /etc/lightdm/lightdm.conf.d/90-mint-docente.conf <<EOF
[Seat:*]
autologin-user=$DOCENTE_USER
autologin-user-timeout=0
EOF

log "Quitar accesos usados solo durante el build"
rm -f /etc/sudoers.d/99-packer /root/postinstalacion.sh /root/postinstalacion.log
rm -f /etc/systemd/system/packer-ssh.service /etc/systemd/system/multi-user.target.wants/packer-ssh.service
systemctl disable ssh.service ssh.socket 2>/dev/null || true
rm -f /etc/ssh/ssh_host_*

log "Comprobar que no queda ninguna credencial de IA ni de GitHub"
rm -rf "$DOCENTE_HOME/.local/share/opencode" "$DOCENTE_HOME/.local/state/opencode" \
       "$DOCENTE_HOME/.config/mint-docente" "$DOCENTE_HOME/.cache/opencode" \
       "$DOCENTE_HOME/.config/ai.opencode.desktop" \
       "$DOCENTE_HOME/.config/gh" "$DOCENTE_HOME/.local/share/keyrings" \
       "$DOCENTE_HOME/.mozilla"   # primer arranque limpio: OpenCode abre su carpeta, Firefox sin pestañas viejas

log "Restos de paquetes, registros e historial"
apt_get autoremove --purge
apt-get clean
rm -rf /var/tmp/mint-docente /var/lib/apt/lists/* /root/.cache "$DOCENTE_HOME/.cache/pip"
rm -f /root/.bash_history "$DOCENTE_HOME/.bash_history"
journalctl --rotate && journalctl --vacuum-time=1s || true
find /var/log -type f -name '*.gz' -delete
find /var/log -type f -exec truncate -s 0 {} +

log "Identidad única por máquina importada"
truncate -s 0 /etc/machine-id
rm -f /var/lib/dbus/machine-id && ln -s /etc/machine-id /var/lib/dbus/machine-id

log "Rellenar con ceros el espacio libre (OVA más pequeña)"
dd if=/dev/zero of=/ZERO bs=4M status=none || true
rm -f /ZERO
sync

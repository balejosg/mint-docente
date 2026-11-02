#!/bin/bash
# OpenCode Desktop v2 (versión fijada), configuración del sistema, instrucciones y skills.
. /tmp/mint-docente/lib.sh

OPENCODE_VERSION=2.2.5
OPENCODE_DEB_SHA256=c827e714220de744c829fdea55540ed814b1372061e7a905fde0978f9f819de8

log "OpenCode Desktop $OPENCODE_VERSION"
tmp=$(mktemp -d)
curl -fsSL -o "$tmp/opencode.deb" \
  "https://github.com/anomalyco/opencode/releases/download/v$OPENCODE_VERSION/opencode-desktop-linux-amd64.deb"
echo "$OPENCODE_DEB_SHA256  $tmp/opencode.deb" | sha256sum -c -
apt_get install "$tmp/opencode.deb"
rm -rf "$tmp"
# CLI incluida en el paquete (la usa el asistente de primer arranque)
ln -sf /opt/OpenCode/resources/opencode-cli /usr/local/bin/opencode

log "Configuración del sistema (/etc/opencode) — prioridad máxima, sin credenciales"
install -D -m644 "$FILES/opencode/opencode.jsonc" /etc/opencode/opencode.jsonc

log "Instrucciones del agente y skills"
install -D -m644 "$FILES/opencode/AGENTS.md" "$SHARE/AGENTS.md"
rm -rf "$SHARE/skills"
cp -r "$FILES/opencode/skills" "$SHARE/skills"
chmod -R a+rX "$SHARE"

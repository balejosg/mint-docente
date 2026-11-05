#!/bin/bash
# Escritorio llave en mano: carpeta de trabajo, lanzadores, autoarranque y asistente.
. /var/tmp/mint-docente/lib.sh || exit 1
E="$FILES/escritorio"

log "Programas del escritorio"
install -D -m644 "$E/mint-docente-ventana-codigo" /usr/local/lib/mint-docente/ventana-codigo
for s in opencode-docente mint-docente-inicio mint-docente-asistente mint-docente-conectar mint-docente-github; do
  install -m755 "$E/$s" "/usr/local/bin/$s"
done

log "Lanzadores en el menú (sustituyen a los de OpenCode)"
install -m644 "$E/opencode-docente.desktop"   /usr/share/applications/opencode-docente.desktop
install -m644 "$E/cambiar-cuenta-ia.desktop"  /usr/share/applications/cambiar-cuenta-ia.desktop
install -m644 "$E/guia-docente.desktop"       /usr/share/applications/guia-docente.desktop
# Solo en el menú: compartir es opcional, no lleva icono en el escritorio.
install -m644 "$E/conectar-github.desktop"    /usr/share/applications/conectar-github.desktop
for f in ai.opencode.desktop.desktop opencode-desktop.desktop; do
  [ -e "/usr/share/applications/$f" ] || continue
  sed -e '/^NoDisplay=/d' -e '/^\[Desktop Entry\]/a NoDisplay=true' "/usr/share/applications/$f" \
    | install -D -o "$DOCENTE_USER" -g "$DOCENTE_USER" -m644 /dev/stdin "$DOCENTE_HOME/.local/share/applications/$f"
done
update-desktop-database /usr/share/applications || true

log "Autoarranque: asistente (primera vez) + OpenCode"
instalar_docente "$E/mint-docente-inicio.desktop" "$DOCENTE_HOME/.config/autostart/mint-docente-inicio.desktop"

log "Guía rápida en PDF"
pandoc "$FILES/GUIA-DOCENTE.md" -o "$SHARE/GUIA-DOCENTE.pdf" --pdf-engine=xelatex \
  -V mainfont="Carlito" -V lang=es
chmod 644 "$SHARE/GUIA-DOCENTE.pdf"

log "Carpeta de trabajo: la que OpenCode Desktop crea y abre en su primer arranque"
# OpenCode Desktop v2 no admite abrir una carpeta concreta: en su primer
# arranque crea y abre «Documentos/<Proyecto predeterminado>». La creamos
# vacía (OpenCode la acepta si ya existe) y el icono «Aula» apunta a ella.
DOCUMENTOS=$(como_docente xdg-user-dir DOCUMENTS 2>/dev/null || true)
[ -n "$DOCUMENTOS" ] && [ "$DOCUMENTOS" != "$DOCENTE_HOME" ] || DOCUMENTOS="$DOCENTE_HOME/Documentos"
TRABAJO="$DOCUMENTOS/Proyecto predeterminado"
como_docente mkdir -p "$TRABAJO"
ln -sfn /media/sf_Intercambio "$DOCENTE_HOME/Intercambio"
chown -h "$DOCENTE_USER:$DOCENTE_USER" "$DOCENTE_HOME/Intercambio"

log "Iconos en el escritorio (se marcan como de confianza al iniciar sesión)"
ESCRITORIO=$(como_docente xdg-user-dir DESKTOP 2>/dev/null || true)
[ -n "$ESCRITORIO" ] && [ "$ESCRITORIO" != "$DOCENTE_HOME" ] || ESCRITORIO="$DOCENTE_HOME/Escritorio"
como_docente mkdir -p "$ESCRITORIO"
como_docente ln -sfn "$TRABAJO" "$ESCRITORIO/Aula"
for f in opencode-docente.desktop guia-docente.desktop cambiar-cuenta-ia.desktop; do
  instalar_docente "/usr/share/applications/$f" "$ESCRITORIO/$f" 755
done

log "Firefox sin páginas de primer uso (el docente lo abre para GitHub)"
install -D -m644 "$FILES/firefox-policies.json" /etc/firefox/policies/policies.json

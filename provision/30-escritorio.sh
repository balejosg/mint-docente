#!/bin/bash
# Escritorio llave en mano: carpeta de trabajo, lanzadores, autoarranque y asistente.
. /var/tmp/mint-docente/lib.sh || exit 1
E="$FILES/escritorio"

log "Programas del escritorio"
for s in opencode-docente mint-docente-inicio mint-docente-asistente mint-docente-conectar; do
  install -m755 "$E/$s" "/usr/local/bin/$s"
done

log "Lanzadores en el menú (sustituyen a los de OpenCode)"
install -m644 "$E/opencode-docente.desktop"   /usr/share/applications/opencode-docente.desktop
install -m644 "$E/cambiar-cuenta-ia.desktop"  /usr/share/applications/cambiar-cuenta-ia.desktop
install -m644 "$E/guia-docente.desktop"       /usr/share/applications/guia-docente.desktop
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

log "Carpeta de trabajo (vacía) y enlace a la carpeta compartida"
como_docente mkdir -p "$DOCENTE_HOME/Aula"
ln -sfn /media/sf_Intercambio "$DOCENTE_HOME/Intercambio"
chown -h "$DOCENTE_USER:$DOCENTE_USER" "$DOCENTE_HOME/Intercambio"

log "Iconos en el escritorio (marcados como de confianza)"
ESCRITORIO=$(como_docente xdg-user-dir DESKTOP 2>/dev/null || true)
[ -n "$ESCRITORIO" ] && [ "$ESCRITORIO" != "$DOCENTE_HOME" ] || ESCRITORIO="$DOCENTE_HOME/Escritorio"
como_docente mkdir -p "$ESCRITORIO"
como_docente ln -sfn "$DOCENTE_HOME/Aula" "$ESCRITORIO/Aula"
for f in opencode-docente.desktop guia-docente.desktop cambiar-cuenta-ia.desktop; do
  instalar_docente "/usr/share/applications/$f" "$ESCRITORIO/$f" 755
  sum=$(sha256sum "$ESCRITORIO/$f" | cut -d' ' -f1)
  como_docente dbus-run-session -- gio set "$ESCRITORIO/$f" metadata::xfce-exe-checksum "$sum" || true
done

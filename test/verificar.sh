#!/bin/bash
# Verificación dentro de la VM al final del build. Si algo falla, el build falla.
set -uo pipefail
U=docente
T=$(sudo -u $U mktemp -d)
FALLOS=0
ok()   { printf '  \033[32m✔\033[0m %s\n' "$1"; }
mal()  { printf '  \033[31m✘\033[0m %s\n' "$1"; FALLOS=$((FALLOS+1)); }
prueba() { local d="$1"; shift; if sudo -u $U -H bash -c "cd '$T' && export PATH=/opt/docente-py/bin:\$PATH && $*" >>"$T/log" 2>&1; then ok "$d"; else mal "$d"; fi; }

echo "== Verificación de Mint Docente"

prueba "OpenCode Desktop instalado"     "test -x /opt/OpenCode/ai.opencode"
prueba "CLI de OpenCode"                "opencode --version"
prueba "Configuración del sistema"      "test -r /etc/opencode/opencode.jsonc"
prueba "Instrucciones (AGENTS.md)"      "test -r /usr/share/mint-docente/AGENTS.md"
prueba "7 skills"                       "test \$(ls /usr/share/mint-docente/skills/*/SKILL.md | wc -l) -eq 7"
prueba "Skill compartir"                "f=/usr/share/mint-docente/skills/compartir/SKILL.md; grep -q \"^name: compartir\" \$f && grep -q \"^description:\" \$f"
prueba "Skill normativa-madrid"        "f=/usr/share/mint-docente/skills/normativa-madrid/SKILL.md; grep -q \"^name: normativa-madrid\" \$f && grep -q \"^description:\" \$f"
prueba "Plantillas docx/odt/typ"        "ls /usr/share/mint-docente/plantillas/{temario,actividad,rubrica,examen}.{docx,odt} /usr/share/mint-docente/plantillas/examen.typ"
prueba "Guía rápida PDF"                "test -s /usr/share/mint-docente/GUIA-DOCENTE.pdf"
prueba "normativa-madrid catalogo"     "normativa-madrid catalogo --etapa primaria | grep -q MAD.EP.MAT.C2"
prueba "Una vista por ficha"           "test \$(ls /usr/share/mint-docente/normativa/*.md | wc -l) -eq \$(normativa-madrid catalogo | python3 -c 'import json,sys; print(len(json.load(sys.stdin)[\"fichas\"]))') && test \$(ls /usr/share/mint-docente/normativa/*.md | wc -l) -gt 0"
prueba "Búsqueda sin tildes"            "normativa-madrid buscar division --en MAD.EP.MAT.C2 | grep -q MAD.EP.MAT.C2"
prueba "Comprobación de códigos"        "printf 'Criterio MAD.EP.MAT.C2.CE.99\n' > cur.md && pandoc cur.md -o cur.docx && { normativa-madrid comprobar cur.docx >/dev/null; test \$? -eq 1; }"

prueba "pandoc md→docx"                 "printf '# Prueba\n\nÁrbol, cigüeña, \$x^2\$\n' > p.md && pandoc p.md -o p.docx"
prueba "LibreOffice docx→pdf"           "soffice -env:UserInstallation=file://$T/lo --headless --convert-to pdf p.docx && test -s p.pdf"
prueba "LibreOffice con Java"           "soffice -env:UserInstallation=file://$T/lo --headless --convert-to odt p.docx && test -s p.odt && java -version"
prueba "pandoc md→pdf (xelatex)"        "pandoc p.md -o p2.pdf --pdf-engine=xelatex && test -s p2.pdf"
prueba "markitdown lee docx"            "markitdown p.docx | grep -q cigüeña"
prueba "python-docx/openpyxl/pptx/pypdf" "python3 -c 'import docx, openpyxl, pptx, pypdf, pdfplumber, odf, reportlab'"
prueba "Typst con fórmulas"             "cp /usr/share/mint-docente/plantillas/examen.typ . && typst compile examen.typ && test -s examen.pdf"
prueba "OCR en español"                 "convert -size 1400x300 xc:white -font Liberation-Serif -pointsize 64 -fill black -annotate +40+160 'Programación didáctica' esc.png && img2pdf esc.png -o esc.pdf && ocrmypdf -q -l spa esc.pdf ocr.pdf && pdftotext ocr.pdf - | grep -qi 'did'"
prueba "Fuentes Carlito y Arial"        "fc-list | grep -q Carlito && fc-list | grep -qi Arial"
prueba "Guest Additions"                "lsmod | grep -q vboxguest"
prueba "docente en vboxsf"              "id -nG | grep -qw vboxsf"
# Sin carpeta compartida en el build, vboxsf no se carga solo: se carga a mano
# para comprobar que el módulo sigue disponible sin compiladores.
modprobe vboxsf 2>/dev/null || true
prueba "Módulo vboxsf"                  "lsmod | grep -qw vboxsf"
# Sin sesión gráfica en el build: no hay proceso VBoxClient --clipboard que buscar.
prueba "Portapapeles de VirtualBox"     "command -v VBoxClient && lsmod | grep -qw vboxguest"
prueba "Kernel retenido"                "apt-mark showhold | grep '^linux-image-' | grep -qvE '[0-9]+\.[0-9]+\.[0-9]+-[0-9]+'"
prueba "Un solo kernel"                 "test \$(ls /boot/vmlinuz-* | wc -l) -eq 1"
prueba "Sin Thunderbird"                "! command -v thunderbird"
prueba "Fuentes para documentos"        "fc-list | grep -q 'Noto Sans' && fc-list | grep -q 'DejaVu Sans' && fc-list | grep -q Carlito && fc-list | grep -qi Arial"
prueba "Locales en español"             "test -d /usr/share/locale/es/LC_MESSAGES && test ! -e /usr/share/locale/ja"
prueba "Carpeta de trabajo vacía"       "D=\$(xdg-user-dir DOCUMENTS)/'Proyecto predeterminado'; test -d \"\$D\" && test -z \"\$(ls -A \"\$D\")\""
prueba "Icono Aula → carpeta de trabajo" "test \"\$(readlink -f \"\$(xdg-user-dir DESKTOP)/Aula\")\" = \"\$(xdg-user-dir DOCUMENTS)/Proyecto predeterminado\""
prueba "Autoarranque configurado"       "test -r ~/.config/autostart/mint-docente-inicio.desktop"
prueba "GitHub CLI"                     "gh --version"
prueba "Sin credenciales de GitHub"     "! gh auth status >/dev/null 2>&1 && test ! -e ~/.config/gh/hosts.yml"
prueba "Lanzador Conectar GitHub"       "test -r /usr/share/applications/conectar-github.desktop"
prueba "Sin icono de GitHub en el escritorio" "test ! -e \"\$(xdg-user-dir DESKTOP)/conectar-github.desktop\""
prueba "Orden compartir sin sesión"     "{ compartir estado >/dev/null 2>&1; test \$? -eq 3; }"
prueba "Pruebas de compartir"           "COMPARTIR_BIN=/usr/local/bin/compartir python3 /var/tmp/mint-docente/test/test_compartir.py"
prueba "Sin credenciales de IA"        "! opencode auth list 2>/dev/null | grep -qiE 'copilot|anthropic|openai|openrouter'"

if [ $FALLOS -gt 0 ]; then
  echo "== $FALLOS prueba(s) fallida(s). Registro:"; tail -n 80 "$T/log"
  exit 1
fi
rm -rf "$T"
echo "== Todo correcto"

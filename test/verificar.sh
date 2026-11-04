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
prueba "5 skills"                       "test \$(ls /usr/share/mint-docente/skills/*/SKILL.md | wc -l) -eq 5"
prueba "Plantillas docx/odt/typ"        "ls /usr/share/mint-docente/plantillas/{temario,actividad,rubrica,examen}.{docx,odt} /usr/share/mint-docente/plantillas/examen.typ"
prueba "Guía rápida PDF"                "test -s /usr/share/mint-docente/GUIA-DOCENTE.pdf"
prueba "CLI curriculo"                  "curriculo areas"
prueba "Vistas del currículo"           "test \$(ls /usr/share/mint-docente/curriculo/*.md | wc -l) -eq \$(python3 -c 'import json; print(len(json.load(open(\"/usr/share/mint-docente/curriculo/primaria-madrid.json\", encoding=\"utf-8\"))[\"areas\"]))')"
prueba "Búsqueda sin tildes"            "curriculo buscar division --area mates"
prueba "Comprobación de códigos"        "printf 'Criterio MAT.9.9.C2\n' > cur.md && pandoc cur.md -o cur.docx && { curriculo comprobar cur.docx; test \$? -eq 1; }"

prueba "pandoc md→docx"                 "printf '# Prueba\n\nÁrbol, cigüeña, \$x^2\$\n' > p.md && pandoc p.md -o p.docx"
prueba "LibreOffice docx→pdf"           "soffice -env:UserInstallation=file://$T/lo --headless --convert-to pdf p.docx && test -s p.pdf"
prueba "pandoc md→pdf (xelatex)"        "pandoc p.md -o p2.pdf --pdf-engine=xelatex && test -s p2.pdf"
prueba "markitdown lee docx"            "markitdown p.docx | grep -q cigüeña"
prueba "python-docx/openpyxl/pptx/pypdf" "python3 -c 'import docx, openpyxl, pptx, pypdf, pdfplumber, odf, reportlab'"
prueba "Typst con fórmulas"             "cp /usr/share/mint-docente/plantillas/examen.typ . && typst compile examen.typ && test -s examen.pdf"
prueba "OCR en español"                 "convert -size 1400x300 xc:white -font Liberation-Serif -pointsize 64 -fill black -annotate +40+160 'Programación didáctica' esc.png && img2pdf esc.png -o esc.pdf && ocrmypdf -q -l spa esc.pdf ocr.pdf && pdftotext ocr.pdf - | grep -qi 'did'"
prueba "Fuentes Carlito y Arial"        "fc-list | grep -q Carlito && fc-list | grep -qi Arial"
prueba "Guest Additions"                "lsmod | grep -q vboxguest"
prueba "docente en vboxsf"              "id -nG | grep -qw vboxsf"
prueba "Carpeta de trabajo vacía"       "D=\$(xdg-user-dir DOCUMENTS)/'Proyecto predeterminado'; test -d \"\$D\" && test -z \"\$(ls -A \"\$D\")\""
prueba "Icono Aula → carpeta de trabajo" "test \"\$(readlink -f \"\$(xdg-user-dir DESKTOP)/Aula\")\" = \"\$(xdg-user-dir DOCUMENTS)/Proyecto predeterminado\""
prueba "Autoarranque configurado"       "test -r ~/.config/autostart/mint-docente-inicio.desktop"
prueba "Sin credenciales de IA"         "! opencode auth list 2>/dev/null | grep -qiE 'copilot|anthropic|openai|openrouter'"

if [ $FALLOS -gt 0 ]; then
  echo "== $FALLOS prueba(s) fallida(s). Registro:"; tail -n 80 "$T/log"
  exit 1
fi
rm -rf "$T"
echo "== Todo correcto"

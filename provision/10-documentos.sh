#!/bin/bash
# Herramientas de documentos para el agente: Office, PDF, OCR, fórmulas y Python.
. /var/tmp/mint-docente/lib.sh || exit 1

TYPST_VERSION=v0.15.1
TYPST_SHA256=a6d077d0a95eed5a2eba715b2dae06be954f624ccbf85758a03f389ded33118c

log "Aceptar la licencia de las fuentes de Microsoft (Arial, Times, Verdana…)"
echo "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" | debconf-set-selections

log "Paquetes de documentos"
apt_get update
apt_get install \
  libreoffice libreoffice-l10n-es libreoffice-help-es \
  pandoc \
  ttf-mscorefonts-installer fonts-crosextra-carlito fonts-crosextra-caladea \
  fonts-liberation fonts-liberation2 fonts-dejavu fonts-noto-core fonts-noto-color-emoji \
  poppler-utils qpdf ghostscript img2pdf pdfarranger \
  ocrmypdf tesseract-ocr tesseract-ocr-spa tesseract-ocr-cat tesseract-ocr-eng \
  tesseract-ocr-glg tesseract-ocr-eus \
  texlive-latex-recommended texlive-latex-extra texlive-xetex texlive-lang-spanish \
  texlive-fonts-recommended lmodern \
  imagemagick git jq zenity xclip unzip p7zip-full curl xournalpp \
  python3-venv python3-pip

log "Typst $TYPST_VERSION"
tmp=$(mktemp -d)
curl -fsSL -o "$tmp/typst.tar.xz" \
  "https://github.com/typst/typst/releases/download/$TYPST_VERSION/typst-x86_64-unknown-linux-musl.tar.xz"
echo "$TYPST_SHA256  $tmp/typst.tar.xz" | sha256sum -c -
tar -xJf "$tmp/typst.tar.xz" -C "$tmp"
install -m755 "$tmp"/typst-x86_64-unknown-linux-musl/typst /usr/local/bin/typst
rm -rf "$tmp"

log "Entorno Python para el agente en /opt/docente-py"
python3 -m venv --system-site-packages /opt/docente-py
/opt/docente-py/bin/pip install --no-cache-dir --upgrade pip
/opt/docente-py/bin/pip install --no-cache-dir \
  python-docx openpyxl python-pptx pypdf pdfplumber odfpy reportlab \
  "markitdown[docx,pdf,pptx,xlsx]"
# Accesos directos que funcionan aunque el PATH no incluya el entorno
ln -sf /opt/docente-py/bin/markitdown /usr/local/bin/markitdown
ln -sf /opt/docente-py/bin/python /usr/local/bin/python-docente

log "Plantillas del sistema (.docx/.odt generadas desde Markdown, .typ)"
P="$SHARE/plantillas"
rm -rf "$P" && mkdir -p "$P"
cp "$FILES/plantillas/examen.typ" "$P/"
pandoc -o "$P/referencia.docx" --print-default-data-file reference.docx
for f in "$FILES"/plantillas/fuentes/*.md; do
  n=$(basename "$f" .md)
  pandoc "$f" -o "$P/$n.docx" --reference-doc="$P/referencia.docx"
done
soffice -env:UserInstallation=file:///tmp/lo-build --headless --convert-to odt --outdir "$P" "$P"/temario.docx "$P"/actividad.docx "$P"/rubrica.docx "$P"/examen.docx
rm -rf /tmp/lo-build
chmod -R a+rX "$SHARE"

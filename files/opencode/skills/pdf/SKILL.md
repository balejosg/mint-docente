---
name: PDF
description: Leer, extraer texto o tablas, unir, dividir, reordenar, rotar, comprimir, pasar OCR a escaneados, convertir imágenes a PDF y rellenar formularios PDF.
---

# PDF

## Leer y extraer
- Texto: `pdftotext -layout archivo.pdf -`, o `markitdown archivo.pdf`.
- Tablas: pdfplumber (`page.extract_tables()`).
- Ver una página: `pdftoppm -png -r 80 -f N -l N archivo.pdf /tmp/pag` y mira la imagen.
- Si `pdftotext` devuelve vacío o basura, es un **escaneado**: aplica OCR.

## OCR (escaneados)
`ocrmypdf -l spa --deskew --rotate-pages entrada.pdf salida.pdf`
- Para varios idiomas: `-l spa+eng` (también disponibles `cat`, `glg` y `eus`).
- Si ya tiene texto: añade `--skip-text` (o `--force-ocr` para rehacerlo).
- Guarda el resultado como `… (OCR).pdf`, sin sobrescribir el original.

## Manipular
- Unir: `pdfunite a.pdf b.pdf salida.pdf`.
- Extraer páginas: `qpdf entrada.pdf --pages . 3-5 -- salida.pdf`.
- Reordenar: `qpdf entrada.pdf --pages . 2,1,3-z -- salida.pdf`.
- Rotar: `qpdf entrada.pdf salida.pdf --rotate=+90:1-3`.
- Comprimir: `gs -sDEVICE=pdfwrite -dPDFSETTINGS=/ebook -o salida.pdf entrada.pdf`.
- Imágenes a PDF: `img2pdf *.jpg -o salida.pdf` (y después OCR si son fotos de texto).
- Documentos de oficina a PDF: `soffice --headless --convert-to pdf archivo.docx`.

## Formularios
pypdf: lee los campos con `reader.get_fields()` y rellénalos con `writer.update_page_form_field_values(...)`. Si el PDF no tiene campos, genera una capa con reportlab y superponla, o díselo al docente.

## Para el docente
Para anotar o firmar a mano un PDF puede usar **Xournal++**, y para reordenar páginas visualmente, **PDF Arranger**. Ambos están en el menú.

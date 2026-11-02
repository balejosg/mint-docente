---
name: Documentos Word, LibreOffice, Excel y PowerPoint
description: Leer, crear, editar y convertir documentos .docx, .odt, .xlsx, .ods, .pptx y .odp conservando el formato; plantillas, estilos, tablas, hojas de calificaciones y presentaciones.
---

# Documentos de oficina

## Leer
- Contenido como texto: `markitdown archivo.docx` (también para .pptx y .xlsx).
- .odt/.ods/.odp: conviértelos primero con `soffice --headless --convert-to docx|xlsx|pptx`, o léelos con odfpy.

## Crear
- **Word (.docx):** usa python-docx. Usa estilos (`Heading 1`, `List Bullet`, `Table Grid`) en vez de formato manual. Para partir de una plantilla: `Document('/usr/share/mint-docente/plantillas/actividad.docx')`, vacía el cuerpo de ejemplo y añade el contenido.
- **Desde Markdown:** `pandoc entrada.md -o salida.docx --reference-doc=/usr/share/mint-docente/plantillas/referencia.docx`. Es rápido para textos largos; las tablas complejas se hacen mejor con python-docx.
- **.odt:** genera el .docx y conviértelo: `soffice --headless --convert-to odt salida.docx`.
- **Excel (.xlsx):** usa openpyxl, con encabezados en negrita, anchos de columna ajustados y **fórmulas reales** (`=PROMEDIO`/`=AVERAGE`, ponderaciones) en lugar de valores calculados, para que el docente pueda cambiar notas. Congela la fila de encabezados.
- **PowerPoint (.pptx):** usa python-pptx, con pocas palabras por diapositiva, títulos claros y una idea por diapositiva.

## Editar un documento del docente
1. Nunca sobrescribas el original: guarda `Nombre (revisado).docx`.
2. Con python-docx, modifica los párrafos o runs existentes para conservar los estilos.
3. Si hay controles de cambios o formato complejo, avisa de que conviene revisarlo en LibreOffice.

## Convertir
- A PDF: `soffice --headless --convert-to pdf --outdir DIR archivo`.
- Lotes: `soffice --headless --convert-to pdf *.docx`.
- Si LibreOffice se queda bloqueado (porque hay una ventana abierta), usa un perfil aparte: `soffice -env:UserInstallation=file:///tmp/lo-agente --headless …`.

## Comprobar
`pdftoppm -png -r 60 salida.pdf /tmp/prev` y mira las imágenes para revisar fuentes, tablas y saltos. Las fuentes de Microsoft (Arial, Times New Roman, Calibri→Carlito) están instaladas.

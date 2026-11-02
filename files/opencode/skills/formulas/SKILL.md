---
name: Fórmulas y documentos matemáticos
description: Crear exámenes, fichas y apuntes con fórmulas matemáticas, físicas o químicas usando Typst (recomendado) o LaTeX, y exportarlos a PDF.
---

# Fórmulas y documentos matemáticos

## Typst (por defecto)
Es rápido y sencillo. Parte de `/usr/share/mint-docente/plantillas/examen.typ`:
```
cp /usr/share/mint-docente/plantillas/examen.typ "Examen Tema 3.typ"
typst compile "Examen Tema 3.typ"    # → Examen Tema 3.pdf
```
Sintaxis útil:
- En línea: `$x^2 + 1$`; en bloque: `$ x = (-b plus.minus sqrt(b^2 - 4a c)) / (2a) $` (con espacios junto a los `$`).
- Fracciones `a/b`, raíces `sqrt(x)`, `root(3, x)`, sumatorio `sum_(i=1)^n`, integral `integral_a^b f(x) dif x`, vectores `vec(1, 2)`, matrices `mat(1, 2; 3, 4)`.
- Química: `$"H"_2"O"$`, `$"CO"_2$`.
- Idioma: `#set text(lang: "es")` (ya está en la plantilla).

## LaTeX (si el docente lo pide o aporta un .tex)
`xelatex -interaction=nonstopmode archivo.tex` (dos pasadas si hay referencias). Paquetes disponibles: amsmath, amssymb, babel-spanish, geometry, tikz, siunitx y otros de texlive-latex-extra.

## Word con fórmulas
Si el docente necesita .docx editable con ecuaciones, escribe en Markdown con `$...$` y convierte: `pandoc entrada.md -o salida.docx --reference-doc=/usr/share/mint-docente/plantillas/referencia.docx`. Pandoc genera ecuaciones nativas de Word (OMML).

## Revisión
Mira siempre el PDF como imagen (`pdftoppm -png -r 80`) para comprobar que las fórmulas se ven bien y no se salen de los márgenes. Si es un examen, entrega también el solucionario.

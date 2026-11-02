// Plantilla de examen con fórmulas (Typst). Compilar: typst compile examen.typ
#set document(title: "Examen")
#set page(paper: "a4", margin: (x: 2cm, y: 2cm),
  footer: context align(center, text(9pt)[Página #counter(page).display() de #counter(page).final().first()]))
#set text(lang: "es", font: ("Carlito", "Liberation Sans"), size: 11pt)
#set par(justify: true)
#set enum(numbering: "1.")

#align(center)[
  #text(16pt, weight: "bold")[Examen — [Unidad o tema]] \
  [Materia] · [Curso] · Modelo A
]

#table(columns: (1fr, auto, auto), stroke: 0.5pt, inset: 6pt,
  [*Nombre:*], [*Fecha:*], [*Nota:*])

*Instrucciones:* Dispones de 50 minutos. Total: 10 puntos.

+ Resuelve la ecuación $x^2 - 5x + 6 = 0$. #h(1fr) *[2 p]*
  #v(3cm)
+ Calcula la derivada de $f(x) = (3x^2 + 1) / (x - 2)$. #h(1fr) *[2 p]*
  #v(3cm)
+ Recuerda la fórmula general:
  $ x = (-b plus.minus sqrt(b^2 - 4 a c)) / (2 a) $

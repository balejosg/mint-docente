---
name: curriculo-madrid
description: Consultar el currículo oficial de Primaria de la Comunidad de Madrid (competencias específicas, criterios de evaluación, contenidos y descriptores del perfil de salida) y comprobar que un documento solo cita códigos que existen
---

# Currículo de Primaria de la Comunidad de Madrid

Fuente: Decreto 61/2022 (anexos I y II), con el Decreto 59/2024 que lo modifica (no cambia los anexos). El enlace entre competencias y descriptores procede del RD 157/2022. Todo está instalado en la máquina: no hace falta Internet. Con `curriculo fuentes` ves las fuentes exactas.

## Cuándo usarla
Siempre que el encargo sea de **Primaria en Madrid** y necesite competencias, criterios de evaluación, contenidos o descriptores: programaciones, unidades de programación, situaciones de aprendizaje, rúbricas o exámenes que citen el currículo. Si la comunidad no es Madrid o la etapa no es Primaria, no la uses.

## Cómo hablar con el docente
Nunca le muestres comandos ni códigos internos sin explicarlos. Dile, por ejemplo, «he consultado el currículo de Madrid» y cita en el documento el criterio con su código y su texto.

## Del curso al ciclo
El currículo se organiza por ciclos: C1 (1º y 2º), C2 (3º y 4º) y C3 (5º y 6º). Si el docente habla de curso («3º»), usa `--curso 3` (da el ciclo C2); un curso fuera de 1–6 es un error. **Al redactar para un curso concreto, pasa siempre `--curso` o `--ciclo`**: algunas competencias tienen una redacción distinta en cada ciclo y sin él podrías citar la equivocada.
- **Educación en Valores Cívicos y Éticos (VCE)** solo se imparte en 5º (ciclo 3). Con otro curso, avisa al docente.
- **Segunda Lengua Extranjera (SLE)** y **Tecnología y Robótica (TYR)** son áreas de etapa: un solo bloque para toda Primaria (ciclo `E`); cualquier curso da `E`.
- Religión no tiene currículo en este decreto: no está.
- Nombres coloquiales valen como área («mates», «lengua», «plástica», «sociales», «naturales», «francés», «valores»…). Si es ambiguo («ciencias», «idioma»), la herramienta lista las opciones: pregunta al docente.

## Formato de los códigos
- Competencia específica: `<ÁREA>.CE<n>`, p. ej. `CS.CE1`.
- Criterio de evaluación: `<ÁREA>.<n>.<m>.<ciclo>`, p. ej. `MAT.3.1.C2` o `TYR.2.4.E`.
- Descriptor del perfil de salida: el oficial, p. ej. `CCL1`, `STEM2`.
- **Contenidos: en Madrid se llaman «contenidos», no «saberes básicos», y no tienen código.** Se citan por bloque y sub-bloque con el texto literal (p. ej. «A. Números y operaciones › Operaciones»).
- Las competencias de Ciencias de la Naturaleza (CN) y Ciencias Sociales (CS) muestran su equivalencia en el RD 157/2022 (p. ej. `CS.CE1` = `CMNSC.CE5`): los libros de texto y otras comunidades usan esa numeración, la de Madrid es distinta.

## Qué debe llevar cada unidad de programación (Orden 130/2023, art. 6.6)
La programación didáctica se hace por área y ciclo. Cada unidad de programación incluye:
1. Competencias específicas y criterios de evaluación que trabaja.
2. Contenidos del ciclo.
3. Actividades y situaciones de aprendizaje.
4. Recursos e instrumentos de evaluación.
5. Criterios de calificación de cada elemento.

Las calificaciones de Primaria son IN (menos de 5), SU (de 5 a menos de 6), BI (de 6 a menos de 7), NT (de 7 a menos de 9) y SB (9 o más). Los documentos oficiales de evaluación se gestionan en Raíces; aquí solo se redactan borradores.

## Flujo de trabajo
1. `curriculo areas` — áreas, ciclos y nombres admitidos.
2. `curriculo ver MAT --curso 3 --tipo relaciones` — competencias con sus descriptores y criterios del ciclo. Otros `--tipo`: `competencias`, `criterios`, `contenidos`, `todo`.
3. `curriculo buscar división --area MAT --curso 3` — localizar contenidos o criterios por tema (admite sin tildes y mayúsculas).
4. `curriculo descriptores [CCL]` — texto de los descriptores de una competencia clave.
5. Redactar el documento citando solo lo que haya devuelto la herramienta.
6. **Antes de entregar:** `curriculo comprobar Programación.docx` (acepta .docx, .xlsx, .pdf, .md… y lee también las tablas). Si sale 0, todo existe. Si sale 1, lista los códigos desconocidos con sugerencias: corrígelos y vuelve a ejecutarlo. Si sale 2, hay un error de uso o de archivo.

Ejemplo real: `MAT.3.1.C2` — «Analizar conjeturas matemáticas sencillas identificando patrones, propiedades y relaciones de forma pautada.» (Matemáticas, 3º–4º). `MAT.9.9.C2` no existe y `comprobar` lo marcaría.

## Cómo citar en el documento
- Pon el **código y el texto literal** que devuelve `curriculo`, sin cambiar palabras.
- Si resumes o adaptas el texto, márcalo como tal («resumen de MAT.3.1.C2») y mantén el código.
- Los contenidos se citan por bloque y sub-bloque, con el texto literal y sin inventar números.
- No cites ninguna competencia, criterio ni descriptor que `curriculo` no haya devuelto.

## Si el docente aporta otra numeración
Si pega criterios de su libro de texto, del currículo estatal o de otra comunidad (p. ej. «CE5 de Conocimiento del Medio»), no mezcles numeraciones. Ejecuta `curriculo comprobar` sobre ellos o búscalos con `curriculo buscar`, propón el equivalente de Madrid y pide confirmación. Conocimiento del Medio se reparte en Madrid entre Ciencias de la Naturaleza y Ciencias Sociales.

## Si algo no está en el currículo
Díselo con claridad: «Ese criterio no figura en el currículo de Primaria de Madrid». Ofrece lo más cercano que sí exista (la herramienta sugiere) o redactarlo como objetivo propio del docente, sin código oficial y marcado como tal. Nunca lo inventes.

## Programa bilingüe (Decreto 59/2024)
En centros con programa bilingüe, Matemáticas, Lengua Castellana y Literatura y los contenidos de Historia de España de Ciencias Sociales no pueden impartirse en lengua extranjera. Tenlo en cuenta al proponer actividades en inglés.

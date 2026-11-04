---
name: Temario y programación didáctica
description: Crear o revisar temarios, programaciones didácticas, unidades didácticas y situaciones de aprendizaje (estructura, objetivos, competencias, criterios de evaluación, saberes, temporalización).
---

# Temario y programación didáctica

## Antes de empezar
Si no lo sabes ya, pregunta en una sola tanda:
1. Etapa, curso y materia (p. ej. «3º ESO, Biología y Geología»).
2. Comunidad autónoma, si quiere ajustarse al currículo oficial. Si la etapa es Primaria y la comunidad es Madrid, carga la skill `curriculo-madrid` y usa sus códigos.
3. Qué necesita: temario breve (índice de unidades), programación completa o una unidad o situación de aprendizaje concreta.
4. Número de sesiones o semanas disponibles, y formato de entrega (por defecto .docx + .pdf).

Si el docente aporta un currículo, una programación anterior o un libro, léelo primero (`markitdown`) y respeta su terminología.

## Estructura recomendada (marco LOMLOE)
Para cada **unidad didáctica / situación de aprendizaje**:
- Título y breve justificación (por qué y para qué).
- Temporalización (sesiones) y trimestre.
- Competencias específicas y descriptores del perfil de salida implicados.
- Criterios de evaluación (numerados como en el currículo si se conocen).
- Saberes básicos.
- Secuencia de actividades: inicio → desarrollo → cierre, con producto final.
- Instrumentos de evaluación (rúbrica, lista de cotejo, prueba escrita…) y peso.
- Atención a la diversidad (medidas DUA: varias formas de representación, de expresión y de implicación).
- Recursos y materiales.

El **temario breve** es una tabla: Unidad | Título | Contenidos clave | Sesiones | Trimestre.

**No inventes códigos de criterios** oficiales. Si no tienes el currículo, redáctalos de forma genérica y avisa al docente de que debe revisarlos con el decreto de su comunidad.

## Producción del documento
1. Parte de `/usr/share/mint-docente/plantillas/temario.docx`, o de la plantilla del docente si aporta una.
2. Genera el documento con python-docx (tablas reales, estilos de Título 1/2, sin formato manual repetido) o con pandoc, usando `--reference-doc` con la plantilla.
3. Convierte a PDF con `soffice --headless --convert-to pdf`, revisa las páginas como imagen y corrige.
4. Resume al docente el número de unidades, el total de sesiones y los avisos (p. ej. criterios a revisar).

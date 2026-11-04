# normativa-madrid: la normativa educativa de la Comunidad de Madrid como fuente de datos para agentes — plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Crear un repositorio nuevo, `normativa-madrid`, independiente de tiza y de Mint Docente. Contendrá la normativa educativa de Madrid (Primaria, ESO, Bachillerato y FP): texto literal, código estable, procedencia de cada elemento y una CLI con salida JSON pensada para agentes. Arranca con el módulo `tiza.normativa` y la Primaria de Mint Docente.

**Architecture:**
- **Paquete Python `normativa_madrid`.** En tiempo de ejecución solo usa la biblioteca estándar. Trae los datos (`datos/normas/*.json` y `datos/fichas/*.json`) y la CLI `normativa-madrid`, que nunca usa la red.
- **Herramientas del mantenedor, en `herramientas/`.** Descargan los boletines, extraen las fichas y comprueban que los textos sean literales. Estas sí tienen dependencias: `pdftotext` y `pdfplumber`.
- **Punto de partida.** El modelo, el cargador, las consultas y los extractores de ESO, Bachillerato y FP salen de tiza (`210d608`, MIT, mismo autor). El extractor de Primaria sale de Mint Docente (`herramientas/extraer-curriculo.py`).
- **Lo que hay que añadir:** la etapa Primaria, el campo `codigo_oficial`, la consulta `comprobar` y la salida en Markdown.

**Tech Stack:**
- Python ≥ 3.11 y `unittest`/`pytest` para las pruebas.
- `uv` para el entorno de desarrollo.
- `pdftotext` (poppler) y `pdfplumber` en las herramientas del mantenedor.

**Spec:** no hay documento de spec. Este plan se apoya en:
- `docs/superpowers/notas/2026-11-04-spike-normativa-comun.md` (spike);
- `docs/superpowers/notas/2026-11-04-spike-curriculo.md` (fuentes de Primaria);
- las decisiones del usuario del 2026-11-04: repositorio nuevo; agnóstico de tiza; propósito, «presentar la normativa de la Comunidad de Madrid como una fuente de datos fiable para agentes»; contenidos numerados, pero marcados como no oficiales; códigos de Primaria al formato común.

Como no hay spec, las decisiones que no cubren estos documentos son provisionales y se anotan en el ledger.

## Contexto: qué existe ya (búsqueda del 2026-11-04)

| Proyecto | Qué cubre | Por qué no basta |
|---|---|---|
| [elprofedelabata/curriculo-lomloe](https://github.com/elprofedelabata/curriculo-lomloe) | JSON por comunidad: Andalucía, Aragón, Asturias, Baleares y Canarias, completas; Cantabria, parcial. Distingue `id` estable de `codigoOficial` | **Madrid está pendiente.** No tiene licencia, así que no se puede reutilizar. No da la página del boletín por elemento. No tiene CLI ni interfaz para agentes. Se creó en octubre de 2026 y sigue activo |
| [meabuerro/profe](https://github.com/meabuerro/profe) | Reales decretos estatales extraídos del XML del BOE (MIT) | Solo el currículo estatal; no recoge el desarrollo de Madrid |
| [noeliarico/lomloe](https://huggingface.co/datasets/noeliarico/lomloe) | Primaria y ESO estatales | Solo estatal |
| corrigiendo.es | Corrección comercial con criterios de Madrid | Cerrado; no publica datos |

Nadie ofrece Madrid con procedencia por elemento ni con una interfaz para agentes. La idea de separar `codigo_oficial` del código estable coincide con la de curriculo-lomloe. Más adelante podría ofrecérseles una exportación de Madrid a su esquema si publican una licencia; queda fuera de este plan.

## Global Constraints

- **Independiente de tiza y de Mint Docente.** Ni en el código, ni en los datos, ni en los mensajes, ni en la documentación aparece «tiza», «Moodle», «aula virtual» ni «Mint Docente». La única excepción es la línea de procedencia del README y del CHANGELOG, que dice de dónde salió el código. La prueba `test_independencia` lo comprueba.
- **Ejecución solo con la biblioteca estándar y sin red.** `pyproject.toml` no declara dependencias en `dependencies`; `pdfplumber` va en el grupo `mantenimiento`. Ninguna consulta abre conexiones.
- **Texto literal del boletín.** Se admiten espacios y saltos de línea normalizados. No se admiten palabras cortadas ni unidas, letras perdidas ni restos de cabecera. Las correcciones de errores oficiales se aplican y quedan anotadas en `via`.
- **Todo elemento tiene `procedencia`:** `norma`, `ubicacion` y `pagina`.
- **Formato del código:** `MAD.<ETAPA>.<MATERIA>[.<CURSO|CICLO>].<TIPO>.<n>[.<m>…]`. Etapas: `EP` (Primaria), `ESO`, `BACH` y los títulos de FP (`DAM`, `CMO`…). En Primaria, el ciclo es `C1`, `C2` o `C3`. Tipos: `CE` (competencia), `CR` (criterio), `CO` (contenidos) y, en FP, los que ya usa tiza. Los descriptores son `MAD.EP.DESC.<CLAVE><n>` (propios de Madrid) y `EST.ESO.DESC.…`/`EST.BACH.DESC.…` (estatales, como en tiza).
- **`codigo_oficial`:** es la numeración tal como la publica el boletín («1.1», «a)», «RA1»…), o `null` si el boletín no numera, como pasa con los contenidos de Madrid. Toda respuesta que incluya un elemento con `codigo_oficial: null` lleva el aviso `NUMERACION_NO_OFICIAL`.
- **Mismo contrato de salida que tiza:** un único JSON en stdout y código de salida 0. Si hay error, `{"error": {"codigo", "detalle", "disponibles"}}` y código de salida 2. `comprobar` devuelve 1 si encuentra códigos desconocidos. Se conservan los códigos de error de tiza (`NORMATIVA_SIN_DATOS`, `CODIGO_DESCONOCIDO`, `CURSO_SIN_DATOS`, `DATOS_INVALIDOS` y `SALIDA_INVALIDA`), con mensajes sin mención a tiza.
- **Cursos académicos:** todas las fichas declaran `cursos_academicos`. Los datos iniciales son de `2026-27`.
- **Idioma:** todo en español: código, mensajes, documentación y commits.
- **Licencia:** MIT para el código. Los textos normativos no tienen derechos de autor (art. 13 de la LPI); se dice así en el README.
- **Commits:** terminan en `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. No se hace push ni se crea el remoto sin preguntar al usuario.

## Review Focus

1. **Un contenido mal troceado no puede pasar en silencio.** Es lo que ocurre con Matemáticas de 1.º de ESO en el spike: un bloque cuyo texto contiene viñetas («‒», «-», «x», «•») o sub-bloques numerados («1. Conteo.»). Lo esperable es que la extracción falle y diga el código del bloque. Prueba: `test_bloque_con_vinetas_es_error` (tarea 6).
2. **Ningún código citable puede desaparecer en la migración de Primaria.** Cada criterio y cada competencia del JSON antiguo debe tener exactamente un elemento nuevo con el mismo texto, el mismo ciclo y el mismo `codigo_oficial`. Prueba: `test_migracion_primaria_sin_perdidas` (tarea 4).
3. **Un agente que consulta un contenido puede creer que su código es oficial.** Lo esperable es que la respuesta lo diga. Prueba: `test_aviso_numeracion_no_oficial` (tarea 2).
4. **`comprobar` sobre un .docx con códigos de tablas, notas al pie o encabezados.** Debe encontrar todos los códigos sin falsos positivos, y un fichero ilegible debe dar salida 2, como en Mint Docente. Pruebas: `test_comprobar_docx_tablas` y `test_comprobar_ilegible` (tarea 5).
5. **Una ficha de Primaria por ciclo frente a VCE (solo 5.º) y SLE/TYR (toda la etapa).** Una consulta con `--curso 5` debe encontrar VCE y el ciclo C3; una con `--curso 2` no debe devolver VCE. Prueba: `test_curso_a_ciclo` (tarea 3).

---

### Task 1: Repositorio nuevo con `tiza.normativa` trasplantado

**Files:**
- Create: `$REPO/` (por defecto, carpeta hermana del repositorio de tiza: `…/datos_replicados/Bruno/normativa-madrid`; el usuario lo confirma antes de empezar).
- Create: `$REPO/pyproject.toml`, `README.md`, `LICENSE`, `CHANGELOG.md`, `AGENTS.md` (instrucciones para quien desarrolle en el repo) y `.gitignore` (`cache/`, `.venv/`).
- Create: `$REPO/src/normativa_madrid/{__init__,modelo,datos,consulta,salida,cli,__main__}.py` y `datos/`, copiados de `tiza/src/tiza/normativa/`.
- Create: `$REPO/herramientas/normativa/…`, copiado de `tiza/herramientas/normativa/`.
- Create: `$REPO/tests/…`, copiado de `tiza/tests/normativa/`.
- Create: `$REPO/tests/test_independencia.py`.

**Interfaces:**
- Produces:
  - CLI `normativa-madrid` con las órdenes de `tiza normativa`, que no cambian: `catalogo`, `ficha <codigo>`, `elemento <codigo>`, `norma "<cita>"`, `buscar "<texto>" [--en <codigo>]` y `programacion <eso|bachillerato|fp>`, todas con `--curso-academico AAAA-AA`.
  - Biblioteca: `normativa_madrid.cargar() -> Indice`, `normativa_madrid.consulta.{catalogo,ficha,elemento,norma,buscar,programacion}` y `normativa_madrid.salida`, con las mismas firmas que en tiza.

- [ ] **Step 1: Copiar el código.** No hay `git filter-repo`. Se copian los ficheros de tiza en `210d608` y en el primer commit se cita ese hash como origen.
- [ ] **Step 2: Cambiar el paquete.** `tiza.normativa` pasa a `normativa_madrid` en imports, `importlib.resources` y `herramientas/normativa/comun.py` (`RUTA_FUENTES` y la ruta de datos).
- [ ] **Step 3: Escribir `cli.py`.** Sale de `tiza/cli.py`: `_anadir_curso`, `_normativa` y `_consulta_normativa`, con la misma salida. Los textos de ayuda de los errores salen de `tiza/ayuda.py`, líneas 126–133, reescritos sin «tiza». Por ejemplo, `DATOS_INVALIDOS` dirá «reinstala o actualiza normativa-madrid».
- [ ] **Step 4: Escribir la prueba de independencia.**

```python
def test_independencia():
    prohibidas = re.compile(r"tiza|moodle|aula virtual|mint docente", re.I)
    for ruta in (RAIZ / "src").rglob("*"):   # y herramientas/, tests/ salvo este fichero, README salvo la línea «Origen:»
        ...
        assert not prohibidas.search(texto), ruta
```

- [ ] **Step 5: Ejecutar las pruebas.** Orden: `uv run pytest -q`. Resultado esperado: pasan todas las pruebas trasplantadas y `test_independencia`.
- [ ] **Step 6: Verificar el comportamiento.**
  - `uv run normativa-madrid ficha MAD.ESO.TECDIG.2 | python -m json.tool` debe devolver los 68 elementos.
  - `uv run normativa-madrid elemento NOEXISTE; echo $?` debe dar `CODIGO_DESCONOCIDO` y salida 2.
- [ ] **Step 7: Hacer el commit.** `git init` y un commit con el mensaje `feat: normativa de Madrid como paquete propio (origen: tiza 210d608)`.

### Task 2: `codigo_oficial` y el aviso de numeración no oficial

**Files:**
- Modify: `src/normativa_madrid/modelo.py`: añadir `Elemento.codigo_oficial: str | None`.
- Modify: `src/normativa_madrid/datos.py`: leerlo. Es obligatorio, y su valor es una cadena o `null`.
- Modify: `src/normativa_madrid/salida.py`: añadirlo a `Fragmento` y los avisos a `Respuesta`.
- Modify: los extractores de `herramientas/normativa/` (`extraer_lomloe`, `extraer_fp` y `extraer_descriptores`) para que lo rellenen.
- Regenerar: `src/normativa_madrid/datos/fichas/*.json` (orden `regenerar`).
- Test: `tests/test_datos.py`, `tests/test_consulta.py` y `tests/test_cli.py`.

**Interfaces:**
- Produces:
  - `Elemento.codigo_oficial`;
  - en la salida, `Fragmento.codigo_oficial` y `Respuesta.avisos: list[{"codigo": str, "detalle": str}]`;
  - el aviso `NUMERACION_NO_OFICIAL`, con el detalle: «Estos elementos no están numerados en el boletín; su código es estable pero no oficial: no lo presentes como numeración oficial».

Valores que pone cada extractor:

| Elemento | Valor |
|---|---|
| Competencia específica | `"<n>"` |
| Criterio | `"<n>.<m>"` |
| Descriptor | `"CCL1"`… |
| Bloque de contenidos | `"A"`… (el boletín pone la letra) |
| Contenido | `null` |
| FP | Lo que imprime el boletín (`"a)"`, `"1."`…). El extractor ya lee esa numeración para construir el código |
| `requisito_programacion` | `null` |

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_codigo_oficial_obligatorio`: una ficha sintética sin el campo da `DATOS_INVALIDOS`, y el detalle nombra `codigo_oficial`.
  - `test_codigo_oficial_de_criterio`: `elemento MAD.ESO.TECDIG.2.CR.1.1` devuelve `codigo_oficial == "1.1"`.
  - `test_aviso_numeracion_no_oficial`:
    - `elemento MAD.ESO.TECDIG.2.CO.A.1` devuelve `codigo_oficial is None` y un aviso cuyo `codigo` es `NUMERACION_NO_OFICIAL`;
    - `elemento MAD.ESO.TECDIG.2.CR.1.1` no devuelve ningún aviso.
- [ ] **Step 2: Comprobar que fallan.** Orden: `uv run pytest -q -k "codigo_oficial or numeracion"`. Resultado esperado: FAIL.
- [ ] **Step 3: Implementar.** Modelo, cargador, salida y extractores. Después, `regenerar --cache <dir>` y `comprobar --cache <dir>`, que debe dar 0 problemas.
- [ ] **Step 4: Comprobar que pasan.** Orden: `uv run pytest -q`. Resultado esperado: PASS.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: codigo_oficial y aviso de numeración no oficial`.

### Task 3: Etapa Primaria en el modelo

**Files:**
- Modify: `modelo.py`:
  - `ETAPAS += ("primaria",)`;
  - nuevo campo `Ficha.ciclo: str | None`;
  - `DATOS_POR_TIPO["competencia_especifica"] = {"equivalencia_rd": str}`.
- Modify: `datos.py`. Validaciones:
  - `ciclo` debe cumplir `C[1-3]`;
  - `ciclo` y `curso` no pueden aparecer juntos;
  - en Primaria, la ficha debe tener `ciclo`, o `curso` (`"5"` en VCE), o ninguno de los dos, que significa toda la etapa (SLE y TYR).
- Modify: `consulta.py` y `cli.py`:
  - `catalogo --etapa primaria`;
  - `ficha` acepta el código de área y `--curso N` o `--ciclo CN` y resuelve la ficha: el curso 1–2 corresponde a C1, el 3–4 a C2 y el 5–6 a C3.
- Test: `tests/test_primaria_modelo.py`, con fichas sintéticas en `tests/datos_sinteticos/`.

**Interfaces:**
- Produces:
  - `consulta.ficha(datos, codigo, *, curso=None, ciclo=None, curso_academico=None)`. Si `codigo` es un área (`MAD.EP.MAT`), resuelve la ficha del ciclo; si es el de una ficha, se comporta como ahora.
  - `consulta.ciclo_de_curso(curso: int) -> str`.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_ciclo_invalido`: `ciclo: "C4"` da `DATOS_INVALIDOS`.
  - `test_ciclo_y_curso_a_la_vez`: da `DATOS_INVALIDOS`.
  - `test_curso_a_ciclo`, con fichas sintéticas `MAD.EP.MAT.C2`, `MAD.EP.MAT.C3`, `MAD.EP.VCE` (curso 5) y `MAD.EP.TYR` (toda la etapa):
    - `ficha MAD.EP.MAT --curso 3` → `MAD.EP.MAT.C2`;
    - `ficha MAD.EP.VCE --curso 5` → `MAD.EP.VCE`;
    - `ficha MAD.EP.VCE --curso 2` → `CURSO_SIN_DATOS`;
    - `ficha MAD.EP.TYR --curso 4` → `MAD.EP.TYR`.
  - `test_equivalencia_rd`: el campo `datos.equivalencia_rd = "CMNSC.CE5"` se carga y aparece en la salida.
- [ ] **Step 2: Comprobar que fallan.** Resultado esperado: FAIL.
- [ ] **Step 3: Implementar.**
- [ ] **Step 4: Comprobar que pasan.** Orden: `uv run pytest -q`. Resultado esperado: PASS.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: etapa Primaria (fichas por ciclo)`.

### Task 4: Extractor de Primaria y migración de los datos de Mint Docente

**Files:**
- Create: `herramientas/normativa/extraer_primaria.py`, adaptado de `mintDocente/herramientas/extraer-curriculo.py`. Se mantiene su lógica de `pdfplumber` y cambia la salida, que pasa a ser fichas.
- Modify: `herramientas/normativa/fuentes.toml`. Añade las normas `bocm-20220718-1` (Decreto 61/2022) y `boe-a-2022-3296` (RD 157/2022), con los sha256 de `docs/superpowers/notas/2026-11-04-spike-curriculo.md`. Si no existe ya, añade también `bocm-20230130-23` (Orden 130/2023). Añade los `recuentos` de cada ficha `MAD.EP.*` a partir de `mintDocente/files/curriculo/recuentos-esperados.json`. `bocm-20240613-2` ya existe.
- Modify: `herramientas/normativa/lote.toml` y `regenerar.py`, para añadir el extractor `primaria`.
- Create: `datos/fichas/MAD.EP.<ÁREA>.<C1|C2|C3>.json` (SLE y TYR sin ciclo; VCE con curso 5) y `MAD.EP.DESC.json`.
- Test:
  - `tests/test_extraer_primaria.py`;
  - `tests/test_migracion_primaria.py`, que lee una copia congelada de `primaria-madrid.json` en `tests/referencia/primaria-madrid-bd89f26.json`.

**Interfaces:**
- Consumes:
  - modelo de Primaria (tarea 3) y `codigo_oficial` (tarea 2);
  - el JSON antiguo: `areas[].{codigo, competencias[{codigo, numero, texto, variantes, descriptores, equivalencia_rd}], criterios[{codigo, ciclo, competencia, texto, pagina}], contenidos[{ciclo, bloque, subbloque, nivel, texto, pagina}]}` y `descriptores`.
- Produces: el mapa de códigos antiguo → nuevo como función `codigo_nuevo(antiguo: str) -> str`, en `herramientas/normativa/extraer_primaria.py`. La usa la tarea 5 para sugerir:
  - `MAT.1.1.C2` → `MAD.EP.MAT.C2.CR.1.1`;
  - `MAT.CE1` → `MAD.EP.MAT.C1.CE.1`, es decir, la competencia del primer ciclo en que aparece.

**Reglas de la migración:**
- **Competencias.** Cada ficha de ciclo lleva todas las competencias del área. El texto es la `variante` de ese ciclo si existe; si no, el `texto`.
- **Relaciones.** `relaciones.descriptores` apunta a `MAD.EP.DESC.*`. La `equivalencia_rd` se copia a `datos`.
- **Contenidos.** Tienen la forma `CO.<bloque>[.<n sub-bloque>].<n>`. El sub-bloque es un `bloque_contenidos` cuyo `padre` es el bloque. Las viñetas de nivel 2 cuelgan de la viñeta anterior de nivel 1.
- **`ubicacion`.** Ejemplos: `"anexo II, Matemáticas, segundo ciclo, criterios"` y `"anexo I, descriptores"`.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_migracion_primaria_sin_perdidas`. Para cada criterio del JSON de referencia, existe exactamente un elemento `codigo_nuevo(c.codigo)` con:
    - el mismo texto, normalizado igual que en `comprobar`;
    - el ciclo de su ficha igual a `c.ciclo`;
    - `codigo_oficial == f"{n}.{m}"`;
    - `procedencia.pagina == c.pagina`.

    Para cada competencia y ciclo, el texto debe coincidir con `variantes.get(ciclo, texto)`. Para cada contenido debe haber un elemento con el mismo texto y la misma página, y el número total de contenidos de cada ficha debe ser igual al número de contenidos de ese ciclo en el JSON antiguo. Además, `len(descriptores)` debe coincidir.
  - `test_recuentos_primaria`: `comprobar_recuentos` no da problemas en ninguna ficha `MAD.EP.*`.
  - `test_contenidos_sin_codigo_oficial`: ningún elemento `contenido` de `MAD.EP.*` tiene `codigo_oficial`.
- [ ] **Step 2: Comprobar que fallan.** Resultado esperado: FAIL, porque las fichas aún no existen.
- [ ] **Step 3: Implementar el extractor y regenerar.** Órdenes: `regenerar --cache <dir>` y luego `comprobar --cache <dir>`. Resultado esperado: 0 problemas. Si la comprobación de literalidad no acepta el texto de `pdfplumber` (puede que trocee distinto que `pdftotext`), compara con el texto de `pdfplumber` de esa página. Anota la decisión en `fuentes.toml`.
- [ ] **Step 4: Comprobar que pasan.** Orden: `uv run pytest -q`. Resultado esperado: PASS.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: Primaria de Madrid (Decreto 61/2022) migrada al formato común`.

### Task 5: Consulta `comprobar`

**Files:**
- Create: `src/normativa_madrid/documentos.py`, con la lectura de .md, .txt, .docx y .odt. Se adapta `_texto_odt` y `leer_texto` de `mintDocente/files/curriculo/curriculo.py`, solo con la biblioteca estándar: `zipfile` y `xml.etree`.
- Modify: `consulta.py`, `salida.py` y `cli.py`, para añadir la orden `comprobar <fichero>… [--curso-academico]`.
- Test: `tests/test_comprobar.py`, con documentos de prueba en `tests/documentos/`. Se generan en la prueba con `zipfile`, sin LibreOffice.

**Interfaces:**
- Produces:
  - `consulta.comprobar(datos, textos: dict[str, str], curso_academico=None) -> salida.Comprobacion`;
  - `salida.Comprobacion`: `{"ficheros": [{"fichero", "conocidos": [codigo], "desconocidos": [{"codigo", "sugerencias": [codigo]}]}], "avisos": [...]}`.
  - Códigos de salida: 0 si todo está limpio, 1 si hay desconocidos y 2 si hay un fichero ilegible o inexistente (`FICHERO_ILEGIBLE`).

**Qué cuenta como código:**
- Se reconoce cualquier token con el patrón `\b(?:MAD|EST)\.[A-Z0-9]+(?:\.[A-Za-z0-9ñ]+)+\b`.
- Si un token empieza por el código de una ficha conocida pero no existe, es desconocido.
- Un `MAD.` seguido de una etapa que no existe es desconocido.
- Los códigos con el formato antiguo de Primaria (`MAT.1.1.C2`) se marcan como desconocidos y llevan como sugerencia el código nuevo (`codigo_nuevo`, tarea 4). Esa función se copia a `src/` como tabla de datos, `datos/equivalencias-primaria-2026.json`.
- Las sugerencias se calculan con `difflib.get_close_matches` dentro de la misma ficha, con un máximo de 3.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_comprobar_limpio` → salida 0.
  - `test_comprobar_desconocido`: `MAD.EP.MAT.C2.CR.9.9` → salida 1, con sugerencias de `MAD.EP.MAT.C2`.
  - `test_comprobar_docx_tablas`: un .docx con un código en una celda, otro en una nota al pie y otro en el encabezado → los tres aparecen en `conocidos`.
  - `test_comprobar_formato_antiguo`: `MAT.1.1.C2` → desconocido, con sugerencia `MAD.EP.MAT.C2.CR.1.1`.
  - `test_comprobar_ilegible`: un .docx corrupto → salida 2 y `FICHERO_ILEGIBLE`.
  - `test_comprobar_sin_falsos_positivos`: «MAD.» a final de frase, URLs y «EST.» en abreviaturas no se cuentan.
- [ ] **Step 2: Comprobar que fallan.** Resultado esperado: FAIL.
- [ ] **Step 3: Implementar.**
- [ ] **Step 4: Comprobar que pasan.** Orden: `uv run pytest -q`. Resultado esperado: PASS.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: comprobar los códigos citados en un documento`.

### Task 6: Extracción a prueba de troceados silenciosos

**Files:**
- Modify: `herramientas/normativa/extraer_lomloe.py`. Las ESO de Matemáticas de 1.º y de Biología de 1.º no se añaden aquí, sino en el plan de datos de ESO; esta tarea solo pone la red de seguridad.
- Modify: `herramientas/normativa/comprobar.py`.
- Test: `tests/test_extraer_lomloe.py` (nuevo o ampliado) y `tests/test_herramienta.py`.

**Interfaces:**
- Produces: un problema nuevo en `comprobar`, `TROCEADO_SOSPECHOSO`, y que el extractor dé error ante el mismo caso.

**Qué es un troceado sospechoso:**
- un `bloque_contenidos` cuyo texto contiene un marcador de viñeta (`‒`, `–`, `-`, `x` o `•` tras un espacio y antes de una mayúscula) o un `\b\d+\.\s+[A-ZÁÉÍÓÚ]` después de la primera frase;
- un elemento de más de 1200 caracteres.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_bloque_con_vinetas_es_error`: el texto sintético que reproduce Matemáticas de 1.º («A. Números y operaciones.\n1. Conteo.\n‒ Estrategias…») hace que el extractor dé error y nombre el bloque `A`.
  - `test_comprobar_detecta_troceado`: una ficha con un bloque de 1968 caracteres da `TROCEADO_SOSPECHOSO`.
  - `test_cabecera_con_mayusculas`: «Competencias Específicas.» se reconoce. Es el caso de Biología.
- [ ] **Step 2: Comprobar que fallan.** Resultado esperado: FAIL.
- [ ] **Step 3: Implementar.** Ajustar los patrones `_COMPETENCIAS`, `_CRITERIOS` y `_CONTENIDOS` (con `re.I` en la palabra «Específicas») y añadir la comprobación.
- [ ] **Step 4: Verificar.**
  - `uv run pytest -q` pasa.
  - `regenerar` y `comprobar` sobre los datos actuales dan 0 problemas, porque las fichas existentes no están troceadas.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: la extracción rechaza contenidos mal troceados`.

### Task 7: Salida en Markdown para personas

**Files:**
- Modify: `cli.py`. Las órdenes `ficha` y `elemento` aceptan `--formato json|md`; el valor por defecto es `json`.
- Create: `src/normativa_madrid/markdown.py`.
- Test: `tests/test_markdown.py`.

**Interfaces:**
- Produces: `markdown.ficha_md(FichaSalida) -> str`, con este formato:
  - título;
  - normas, con enlace al PDF;
  - competencias, cada una con sus descriptores y debajo sus criterios;
  - contenidos por bloque y sub-bloque, sin códigos no oficiales a la vista. El código va en un comentario HTML, `<!-- MAD.EP.… -->`, para que un agente que lea la vista pueda citarlo.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_md_criterio_con_numero_oficial`: la vista de `MAD.ESO.TECDIG.2` contiene `"**1.1** "` seguido del texto del criterio.
  - `test_md_contenidos_sin_codigo_visible`: ninguna línea visible contiene `.CO.`; sí aparece en un comentario `<!--`.
- [ ] **Step 2: Comprobar que fallan.** Resultado esperado: FAIL.
- [ ] **Step 3: Implementar.**
- [ ] **Step 4: Comprobar que pasan.** Resultado esperado: PASS.
- [ ] **Step 5: Hacer el commit.** Mensaje: `feat: vista en Markdown de una ficha`.

### Task 8: Contrato para agentes, documentación y paquete instalable

**Files:**
- Create: `docs/contrato.md`, que recoge:
  - cada orden, con un ejemplo de su salida JSON;
  - todos los códigos de error y de aviso;
  - qué garantiza el paquete (texto literal, procedencia, `codigo_oficial`) y qué no garantiza (vigencia posterior a `revisada` y recursos judiciales: hay un contencioso contra el Decreto 61/2022 cuyo resultado no consta).
- Create: `docs/mantener-los-datos.md`, que sale de `herramientas/normativa/README.md` de tiza y añade Primaria y la revisión externa por materia. Ahí se copia el prompt de `mintDocente/docs/superpowers/prompts/revision-curriculo.md`, generalizado a etapas.
- Create: `skill/normativa-madrid/SKILL.md`, una skill genérica para cualquier agente:
  - cuándo consultarla;
  - que no cite nada que la consulta no devuelva;
  - que pase `comprobar` antes de entregar;
  - que no presente como oficial ningún código con `NUMERACION_NO_OFICIAL`;
  - la terminología de Madrid («contenidos»);
  - el art. 9.1 del Decreto 61/2022 en su redacción del Decreto 59/2024 (programa bilingüe).

  Se basa en `mintDocente/files/opencode/skills/curriculo-madrid/SKILL.md`, sin mencionar ningún producto.
- Modify: `README.md` (propósito, cobertura por etapa y materia, instalación, ejemplo para agentes y origen) y `CHANGELOG.md` (0.1.0).
- Modify: `pyproject.toml`, para que `uv build` genere una rueda con los datos incluidos.

- [ ] **Step 1: Escribir la prueba que falla.** `test_contrato_al_dia`: cada código de error y de aviso que aparece en `salida.py` y `ayuda` está documentado en `docs/contrato.md`, y al revés.
- [ ] **Step 2: Escribir la documentación y la skill.** Con eso, la prueba pasa.
- [ ] **Step 3: Verificar el paquete fuera del repositorio.**
  - `uv build`.
  - En un venv nuevo, `pip install dist/normativa_madrid-0.1.0-py3-none-any.whl && normativa-madrid catalogo --etapa primaria`. Resultado esperado: un JSON con las fichas `MAD.EP.*`, sin red.
- [ ] **Step 4: Hacer el commit.** Mensaje: `docs: contrato para agentes, skill y paquete 0.1.0`.

---

## Planes siguientes (fuera de este plan; cada uno con su propio documento)

1. **tiza usa `normativa-madrid`.** Se hace en el repositorio de tiza.
   - `tiza normativa …` llama a `normativa_madrid.consulta` y conserva su salida. Si hay diferencias, como `codigo_oficial` y `avisos`, se documentan en el CHANGELOG de tiza.
   - Se borran `src/tiza/normativa/`, `herramientas/normativa/` y sus pruebas.
   - Las guías de la skill de tiza (`programacion.md`, `unidad.md`, `realinear.md`, `calificador.md`, `dua.md` y `licencias.md`) remiten a los códigos con el aviso `NUMERACION_NO_OFICIAL`.
   - `aislamiento.py` permite `normativa-madrid` al agente.
2. **Mint Docente usa `normativa-madrid`.**
   - La rueda se instala sin red durante el build.
   - Desaparecen `files/curriculo/`, `herramientas/extraer-curriculo.py`, `test/curriculo/` y la CLI `curriculo`, o queda un alias de una línea si conviene.
   - La skill `curriculo-madrid` pasa a ser la skill genérica de la tarea 8.
   - Se actualizan `AGENTS.md`, la skill `temario` (que deja de decir que no hay ESO ni Bachillerato), `GUIA-DOCENTE.md` y `test/verificar.sh`.
   - Hay que hacerlo **antes** de distribuir la OVA de Primaria.
   - tiza se instala en la VM como bloque aparte.
3. **ESO y Bachillerato completos.** Se procesa una materia por ficha:
   - recuentos de la fase 1 del revisor externo;
   - extracción;
   - `comprobar` sin problemas;
   - revisión de la fase 2.

   Se empieza por las troncales de ESO (Matemáticas, Lengua, Inglés, Geografía e Historia, Biología y Física y Química) y se corrigen los casos de Matemáticas de 1.º y Biología de 1.º. Son unas 60 fichas de ESO más Bachillerato.
4. **Normas transversales.** Evaluación de Primaria (Orden 130/2023), las equivalentes de ESO y Bachillerato, y la atención a la diversidad. Con eso se alimentan `programacion primaria` y los bloques pendientes de Mint Docente.
5. **Opcional:** un servidor MCP de solo lectura sobre las mismas consultas, y la exportación al esquema de curriculo-lomloe si publican una licencia.

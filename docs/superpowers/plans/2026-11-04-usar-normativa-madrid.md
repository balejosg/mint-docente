# tiza, tiza-privado y Mint Docente usan normativa-madrid — plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Que los tres proyectos tomen la normativa del paquete `normativa-madrid` 0.1.0 publicado, en vez de llevar cada uno su propia copia.

**Architecture:**
- **tiza** declara la dependencia del wheel de la release y deja de traer `tiza.normativa` y sus herramientas. La orden `tiza normativa …` pasa a ser una envoltura fina de `normativa_madrid.cli.main`: `uv tool install` no expone en el PATH los ejecutables de las dependencias, y además así se mantiene la orden que ya conocen la skill y los docentes.
- **tiza-privado** recibe los mismos commits, adaptados a su nombre de orden y a su skill.
- **Mint Docente** instala el wheel en `/opt/docente-py` (descarga con sha256 fijado, como hace con OpenCode), publica la orden `normativa-madrid` y unas vistas en Markdown, sustituye la skill `curriculo-madrid` por la skill del paquete y elimina su copia propia del currículo.

**Tech Stack:** Python ≥ 3.11, `uv`, pytest (tiza), bash + Packer (Mint).

**Spec:**
- Plan anterior: `docs/superpowers/plans/2026-11-04-normativa-madrid.md`, sección «Planes siguientes», puntos 1 y 2.
- Contrato del paquete: `docs/contrato.md` del repositorio `balejosg/normativa-madrid` en la etiqueta `v0.1.0`.
- Decisiones del usuario del 2026-11-04:
  - el paquete se publica como release v0.1.0 con su wheel;
  - los cambios de tiza, tiza-privado y Mint se quedan en local, sin push ni release, hasta que el usuario lo diga.

## Global Constraints

- **Dependencia, siempre la misma URL:** `normativa-madrid @ https://github.com/balejosg/normativa-madrid/releases/download/v0.1.0/normativa_madrid-0.1.0-py3-none-any.whl`. En Mint, además, con el sha256 del wheel publicado; el controlador lo pasa en el dispatch.
- **Ninguno de los tres repos conserva una copia de los datos ni del código de normativa.** La única fuente es el paquete.
- **El contrato de la salida es el del paquete.** En tiza, `tiza normativa <args>` debe producir exactamente lo mismo que `normativa-madrid <args>`: stdout, código de salida y errores.
- **Repos:**
  - tiza: `/srv/dev-disk-by-uuid-74dc40f5-117d-4922-a32f-0d66e888f8c6/datos_replicados/Bruno/tiza`, rama `feat/normativa-madrid`.
  - tiza-privado: `…/Bruno/tiza-privado`, rama `feat/normativa-madrid`.
  - Mint: `/datos_nvme/run0/mintDocente`, rama `feat/normativa-madrid`.
- **No se hace push en ningún repo ni se crea ninguna release de tiza.** En tiza-privado, `upstream` (el público) es solo de lectura: nunca se sube nada a él.
- **Commits** en español, terminando en `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Versión:** tiza y tiza-privado pasan a `0.40.0`, con entrada en el CHANGELOG.
- **Mint:** la OVA de Primaria aún no se ha distribuido. Los códigos nuevos (`MAD.EP.…`) sustituyen a los antiguos sin alias.

## Review Focus

1. **Un agente que ya usaba `tiza normativa ficha MAD.ESO.TECDIG.2` debe recibir los mismos códigos y textos que antes**, más los campos nuevos (`codigo_oficial`, `avisos`). Prueba: `test_normativa_igual_que_el_paquete` (tarea 1).
2. **`tiza actualizar` y los instaladores deben seguir instalando sin git y sin construir nada.** La dependencia es un wheel, no un repositorio. Prueba: el paso de verificación de la tarea 1, que instala el wheel de tiza en un entorno nuevo con `uv tool install` y ejecuta `tiza normativa catalogo`.
3. **En tiza-privado, la orden es `tiza-privado normativa`.** Ningún texto suyo debe remitir a `tiza normativa` ni al repositorio público. Prueba: `test_normativa_igual_que_el_paquete` adaptado (tarea 3).
4. **En la VM, el agente tiene que encontrar la orden y la skill sin red.** Prueba: `test/verificar.sh` con `normativa-madrid catalogo --etapa primaria` y `comprobar` (tarea 4); la construcción de la OVA la ejecuta la tarea 5.
5. **Un docente de ESO no debe leer en la guía que «no hay currículo de ESO»** cuando sí hay algunas materias. Lo que digan la guía y la skill `temario` debe coincidir con `catalogo`. Revisión de la tarea 4.

---

### Task 1: tiza depende de normativa-madrid

**Files:**
- Modify: `pyproject.toml` (dependencia; dejar de empaquetar los datos de normativa) y `uv.lock`.
- Modify: `src/tiza/cli.py`. El subparser `normativa` pasa a ser `add_parser("normativa", add_help=False)` con un resto de argumentos (`argparse.REMAINDER`) que se entrega a `normativa_madrid.cli.main(argv)`. Su código de salida se devuelve tal cual. Se eliminan `_anadir_curso`, `_normativa` y `_consulta_normativa` si ya no se usan en otro sitio.
- Modify: `src/tiza/ayuda.py`. Se eliminan las entradas de errores que ahora son del paquete (`DATOS_INVALIDOS`, `NORMATIVA_SIN_DATOS`, `CODIGO_DESCONOCIDO`, `CURSO_SIN_DATOS`, `SALIDA_INVALIDA` y las demás que solo use la normativa), salvo que otra orden de tiza las use. Hay que comprobarlo.
- Delete: `src/tiza/normativa/`, `herramientas/normativa/` y `tests/normativa/`.
- Modify: los tests que importaban `tiza.normativa` (`tests/test_skill.py`, `tests/test_ayuda.py` y los que aparezcan).
- Test: `tests/test_cli_normativa.py` (nuevo, corto).

**Interfaces:**
- Consumes: `normativa_madrid.cli.main(argv: list[str] | None) -> int` (paquete 0.1.0).
- Produces: `tiza normativa …`, que el agente sigue usando y que la skill documenta en la tarea 2.

- [ ] **Step 1: Escribir las pruebas que fallan.**
  - `test_normativa_igual_que_el_paquete`, parametrizada con:
    - `catalogo`
    - `ficha MAD.ESO.TECDIG.2`
    - `ficha MAD.EP.MAT --curso 3`
    - `elemento MAD.EP.MAT.C2.CR.1.1`
    - `elemento NOEXISTE`
    - `buscar fracciones`
    - `programacion fp`
    - `ficha MAD.EP.MAT.C2 --formato md`

    Para cada una, comparar stdout y código de salida entre `tiza.cli.main(["normativa", *args])` y `normativa_madrid.cli.main(args)`.
  - `test_normativa_comprobar_documento`: un `.md` con un código conocido y uno desconocido da salida 1 y el mismo documento que el paquete.
  - `test_tiza_no_trae_normativa`: `importlib.util.find_spec("tiza.normativa") is None`, y el wheel construido no contiene `tiza/normativa/`.
- [ ] **Step 2: Comprobar que fallan.** Orden: `uv run pytest -q tests/test_cli_normativa.py`. Resultado esperado: FAIL.
- [ ] **Step 3: Implementar y borrar la copia.** `uv lock` con la URL de la dependencia.
- [ ] **Step 4: Ejecutar la suite completa.** Orden: `uv run pytest -q`. Debe pasar.
- [ ] **Step 5: Verificar la instalación como la hace el docente.**
  - `uv build --wheel`.
  - `UV_TOOL_DIR=$(mktemp -d) UV_TOOL_BIN_DIR=$(mktemp -d) uv tool install dist/tiza-*.whl`.
  - Con ese binario, `tiza normativa catalogo --etapa primaria`. Resultado esperado: un JSON con 25 fichas `MAD.EP.*`.
- [ ] **Step 6: Commit.** Mensaje: `feat: la normativa sale del paquete normativa-madrid`.

### Task 2: skill, guías y documentación de tiza

**Files:**
- Modify: `src/tiza/skill/tiza/SKILL.md` y las guías que citan la normativa: `guias/programacion.md`, `unidad.md`, `realinear.md`, `calificador.md`, `dua.md` y `licencias.md`.
- Modify: `docs/guia.md` (sección «Normativa»), `docs/del-curriculo-al-aula.md`, `README.md`, `CHANGELOG.md` (0.40.0) y `pyproject.toml` (versión 0.40.0).
- Test: los que ya comprueben la skill (`tests/test_skill.py`) y la prueba nueva `test_skill_sin_ordenes_inexistentes`.

**Interfaces:**
- Consumes: el contrato del paquete. El texto de las guías se basa en `docs/contrato.md` y en `skill/normativa-madrid/SKILL.md` de `normativa-madrid@v0.1.0`; se adapta, no se copia entero.

**Qué deben decir la skill y las guías:**
- **Cobertura:** Primaria completa, las materias de ESO y Bachillerato que da `catalogo` y FP DAM. Lo demás, «no está: pide el PDF».
- **`codigo_oficial` y `NUMERACION_NO_OFICIAL`:** un contenido se cita por su texto y su bloque. Su código es estable pero no oficial, sirve para enlazar (por ejemplo, en el calificador) y no se presenta como numeración oficial.
- **Comprobar antes de publicar:** `tiza normativa comprobar <fichero>` antes de publicar cualquier documento que cite códigos.
- **Origen de los datos:** la normativa viene del paquete `normativa-madrid` (se nombra con el enlace del repositorio).

- [ ] **Step 1: Escribir la prueba que falla.** `test_skill_sin_ordenes_inexistentes`: toda orden `tiza normativa <consulta>` citada en la skill o en las guías es una de las de `normativa-madrid --help` (`catalogo`, `ficha`, `elemento`, `norma`, `buscar`, `programacion`, `comprobar`).
- [ ] **Step 2: Reescribir los textos.** Con eso, la prueba y `tests/test_skill.py` pasan.
- [ ] **Step 3: Ejecutar la suite.** Orden: `uv run pytest -q`.
- [ ] **Step 4: Commit.** Mensaje: `docs: la skill y las guías usan la normativa del paquete; versión 0.40.0`.

### Task 3: tiza-privado recibe los mismos cambios

**Files:** los mismos que en las tareas 1 y 2, en `…/Bruno/tiza-privado`. Allí la orden es `tiza-privado` y la skill es `tiza-privado`, con sus propias guías.

- [ ] **Step 1: Traer los commits.** `git fetch upstream`, sin hacer push nunca. Traer los commits de las tareas 1 y 2 con `git cherry-pick` desde el repositorio local de tiza (`git fetch <ruta-a-tiza> feat/normativa-madrid`) y resolver los conflictos con el estilo del privado. Revisar los ficheros que solo existen en el privado y citan la normativa: `grep -rn "normativa" src tests docs README.md AGENTS.md`.
- [ ] **Step 2: Adaptar los nombres.** Orden `tiza-privado normativa` y su skill. La prueba `test_normativa_igual_que_el_paquete` llama a `tiza.cli.main(["normativa", …])`. Hay que comprobar que en la ayuda y en la skill del privado no queda `tiza normativa` ni ninguna mención al repositorio público.
- [ ] **Step 3: Ejecutar la suite y verificar la instalación.** Orden: `uv run pytest -q`, más el paso 5 de la tarea 1 con el wheel del privado.
- [ ] **Step 4: Commits.** Los mismos asuntos que en tiza, como hace el privado.

### Task 4: Mint Docente instala normativa-madrid

**Files:**
- Modify: `provision/20-opencode.sh`. El bloque «Currículo de Primaria (Madrid)» pasa a ser «Normativa de Madrid (normativa-madrid 0.1.0)»:
  - descarga el wheel de la URL de la restricción global y comprueba su sha256;
  - `/opt/docente-py/bin/pip install --no-cache-dir <wheel>`;
  - enlace `/usr/local/bin/normativa-madrid` → `/opt/docente-py/bin/normativa-madrid`;
  - vistas `$SHARE/normativa/<nombre de la ficha>.md`, generadas con `normativa-madrid ficha <código> --formato md` para cada ficha del `catalogo` (todas las etapas);
  - se mantienen las variables para probar el bloque sin ser root (`NORMATIVA_DIR`, `NORMATIVA_BIN`, `NORMATIVA_PY`).
- Delete: `files/curriculo/`, `herramientas/extraer-curriculo.py`, `test/curriculo/` y `requirements-dev.txt` (si solo servía para esto), además de lo que `build.sh` ejecutaba de `test/curriculo`.
- Create: `files/opencode/skills/normativa-madrid/SKILL.md`, copia de `skill/normativa-madrid/SKILL.md` de `normativa-madrid@v0.1.0`. Las adaptaciones al entorno de la VM, como la ruta de las vistas, van en `AGENTS.md`, no en la copia.
- Create: `files/opencode/skills/normativa-madrid/ORIGEN`, con la URL del fichero en la etiqueta y su sha256.
- Delete: `files/opencode/skills/curriculo-madrid/`.
- Modify:
  - `files/opencode/AGENTS.md`: orden, ruta de las vistas y reglas de cita (sin decir «Primaria» como única etapa);
  - `files/opencode/skills/temario/SKILL.md`: cargar `normativa-madrid` en cualquier etapa de Madrid y quitar el aviso de que ESO y Bachillerato no están;
  - `files/GUIA-DOCENTE.md`: sección «Currículo de Madrid» con la cobertura real;
  - `README.md`: tabla, estructura y mantenimiento, ahora en el repositorio del paquete.
- Modify: `test/verificar.sh`. Se sustituyen las pruebas de `curriculo` por:
  - `normativa-madrid catalogo --etapa primaria`;
  - una vista por ficha del catálogo;
  - `buscar division --en MAD.EP.MAT.C2`;
  - `comprobar` sobre un .docx con un código inexistente, que da salida 1;
  - la skill con su frontmatter.
- Modify: `build.sh`. En el anfitrión, antes de Packer, comprueba que `files/opencode/skills/normativa-madrid/SKILL.md` tiene el sha256 de `ORIGEN`.

- [ ] **Step 1: Probar el bloque de aprovisionamiento sin VM.** Con las variables apuntando a un directorio temporal y un venv temporal como `NORMATIVA_PY`: el script descarga, verifica, instala y genera tantas vistas como fichas tiene el catálogo. Una prueba con un sha256 incorrecto debe hacer fallar el bloque.
- [ ] **Step 2: Comprobar el resto en el anfitrión.** `bash -n` en los scripts tocados. `grep -rn "curriculo" files provision test build.sh README.md` solo encuentra la palabra en prosa, no órdenes ni rutas antiguas.
- [ ] **Step 3: Commit.** Mensaje: `feat: la VM usa normativa-madrid en lugar de su copia del currículo`.

### Task 5: Construcción de la OVA y verificación en la VM

- [ ] **Step 1: Construir.** `./build.sh` tarda unos 70 minutos; se lanza en segundo plano. Resultado esperado: termina con `== Todo correcto` y pasan las pruebas nuevas de `test/verificar.sh`.
- [ ] **Step 2: Registrar el resultado.** Si falla, se diagnostica con superpowers:systematic-debugging, se corrige en la tarea 4 y se vuelve a construir.

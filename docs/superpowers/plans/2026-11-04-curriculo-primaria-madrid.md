# Currículo de Primaria de la Comunidad de Madrid en local — Plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Objetivo:** que el agente de Mint Docente trabaje la documentación curricular de Primaria con el currículo **oficial de Madrid**, consultado en local, y que no pueda colar un criterio, una competencia o un descriptor que no exista.

**Arquitectura:** el currículo se extrae **una sola vez, en el anfitrión de desarrollo**, de los PDF del BOCM/BOE a un JSON versionado en el repo (`files/curriculo/primaria-madrid.json`). Otro LLM lo revisa contra el boletín, área por área, y se protege con pruebas de integridad y recuentos. Dentro de la VM, la CLI `curriculo` consulta ese JSON (áreas, competencias, criterios, contenidos y descriptores), busca texto y **comprueba** los códigos citados en un documento. La skill `curriculo-madrid` y una regla en `AGENTS.md` obligan al agente a usarla. El build no descarga ni interpreta nada del BOCM: así es reproducible y no hay red de la que depender.

**Tecnologías:** Python 3 (`/opt/docente-py`, solo biblioteca estándar más `markitdown`, que ya está instalado), `pdfplumber`/`pdftotext` (solo en el extractor del anfitrión), `unittest`, Bash (aprovisionamiento y `test/verificar.sh`).

**Especificación:** este plan parte de lo acordado en la conversación del 2026-11-04, sin documento aparte:
- La máquina es para que docentes de **Primaria de Madrid** creen material y trabajen la **documentación curricular y programática**.
- El bloque 1 es el currículo de Madrid en local, con la normativa de evaluación y el RD 157/2022 (perfil de salida y descriptores) y una tabla de relaciones competencia → criterios → saberes → descriptores. *(Tras el spike: Madrid no vincula los contenidos con las competencias, sino que los da por ciclo; la relación queda en competencia → descriptores + criterios del ciclo, con los contenidos del ciclo aparte.)*
- Regla: el agente **solo cita códigos que existan en esos archivos**.

**Spike hecho (tarea 0):** `docs/superpowers/notas/2026-11-04-spike-curriculo.md`. Tiene las fuentes con su sha256, la estructura del decreto y la estrategia de extracción. Quien implemente **debe leerla**.

## Restricciones globales

- Normas, ya confirmadas en el spike:
  - el Decreto 61/2022 (BOCM 18-07-2022), que es la **única fuente** del Anexo I (descriptores) y del Anexo II (competencias, criterios y contenidos);
  - el Decreto 59/2024, que no cambia los anexos pero se cita;
  - el RD 157/2022, que **solo** se usa para el enlace competencia → descriptores;
  - la Orden 130/2023, que se cita en la skill.

  Las URL y los sha256 están en la nota del spike.
- Áreas y prefijos:
  - `CN` (Ciencias de la Naturaleza), `CS` (Ciencias Sociales), `EA` (Educación Artística), `EF` (Educación Física), `LCL` (Lengua Castellana y Literatura), `ING` (Lengua Extranjera: Inglés) y `MAT` (Matemáticas), con ciclos 1, 2 y 3;
  - `VCE` (Educación en Valores Cívicos y Éticos), solo en 5º, ciclo 3;
  - `SLE` (Segunda Lengua Extranjera) y `TYR` (Tecnología y Robótica), con un solo bloque para toda la etapa, ciclo `E`.
  - Religión queda fuera.
- **Los contenidos no tienen numeración oficial y no llevan código.** Se citan por bloque y sub-bloque con el texto literal.
- Ruta en la VM: `/usr/share/mint-docente/curriculo/` (ya está permitida en `opencode.jsonc` por `/usr/share/mint-docente/*`).
- Sin red en el build para el currículo: el JSON va en el repo y lleva la URL y el sha256 de cada PDF de origen.
- Todo el texto que ve el docente va en español y sin jerga. Los mensajes de la CLI se dirigen al agente, pero también en español.
- Los textos normativos no tienen derechos de autor (art. 13 del TRLPI), así que pueden redistribuirse en la imagen.
- Formato de los códigos, **fijo en todo el plan**. Son códigos de Mint Docente construidos con la numeración oficial:
  - competencia específica: `<ÁREA>.CE<n>`, por ejemplo `CS.CE1`;
  - criterio: `<ÁREA>.<n>.<m>.<ciclo>`, donde el ciclo es `C1`, `C2`, `C3` o `E`. Por ejemplo, `MAT.3.1.C2` o `TYR.2.4.E`;
  - descriptor: el oficial, por ejemplo `CCL1` o `STEM2`.
- Las competencias de CN y CS guardan su equivalente del RD (`equivalencia_rd`, por ejemplo `CS.CE1` → `CMNSC.CE5`), porque los libros de texto y otras comunidades usan la numeración del RD.
- La CLI sale con código 0 si todo va bien, 1 si hay códigos desconocidos o no hay resultados y 2 si el uso es incorrecto.
- No se toca nada de la parte de secundaria salvo `AGENTS.md`, para preguntar la etapa.

## Puntos de revisión

Son los casos que la especificación implica y es más probable que fallen con un docente real. Cada uno tiene su prueba en la tarea responsable.

1. **El docente pega un criterio del currículo estatal o de otra comunidad** (otra numeración, o un criterio que en Madrid no existe). El caso típico es Ciencias Sociales citada como «CE5 de Conocimiento del Medio». `curriculo comprobar` debe marcarlo como desconocido y, si hay equivalencia, proponer la de Madrid → tarea 3.
2. **Se escribe sin tildes o con otras mayúsculas** («division», «LECTURA»). `buscar` debe encontrar «división» y «lectura» → tarea 2.
3. **Los códigos van dentro de tablas de un .docx o de una hoja .xlsx**, que es lo normal en una programación. `comprobar` debe verlos igual que en el texto corrido → tarea 3.
4. **El docente habla de curso y no de ciclo** («3º»). `--curso 3` debe dar el ciclo 2, y un curso fuera de 1–6 debe dar error de uso. En las áreas de etapa (SLE, TYR), cualquier curso devuelve `E`. En VCE, un curso distinto de 5º avisa de que el área solo se imparte en 5º → tarea 2.
5. **Nombres coloquiales de las áreas.** «mates», «lengua», «inglés», «plástica», «música» (las dos → EA), «sociales», «naturales», «educación física», «robótica», «francés» (→ SLE) y «valores» deben resolverse a su área. Uno ambiguo («ciencias», «idioma») debe listar las opciones → tarea 2.

---

### Tarea 0: Spike de fuentes — HECHO

Los resultados están en `docs/superpowers/notas/2026-11-04-spike-curriculo.md`. Ya están incorporados en las restricciones globales y en las tareas 1 a 3.

---

### Tarea 1: Datos del currículo y su integridad

**Archivos:**
- Crear: `herramientas/extraer-curriculo.py`. Solo se usa en el anfitrión y no entra en la imagen.
- Crear: `requirements-dev.txt`, con versiones fijadas de pdfplumber, markitdown[docx,pdf,xlsx], python-docx y openpyxl. Añadir `.venv-dev/` y `cache/fuentes/` a `.gitignore`.
- Crear: `files/curriculo/primaria-madrid.json`, generado por el extractor, revisado y versionado.
- Crear: `files/curriculo/recuentos-esperados.json`, con los recuentos contados en el boletín por alguien distinto del extractor (ver el paso de revisión).
- Crear: `test/curriculo/test_datos.py`.

**Interfaces que produce, el esquema JSON que leen las tareas 2 y 3:**
- Raíz:
  - `version` (fecha de extracción, AAAA-MM-DD);
  - `fuentes`: lista de `{norma, boletin, fecha, url, sha256}`, con las cuatro normas de la nota del spike;
  - `descriptores`: lista de `{codigo, competencia_clave, texto}`, sacados del Anexo I del Decreto 61/2022;
  - `areas`: lista.
- Cada área:
  - `{codigo, nombre, alias: [str], ciclos: ["C1","C2","C3"] | ["C3"] | ["E"]}`;
  - `competencias`: lista de `{codigo, numero, texto, descriptores: [codigo], equivalencia_rd: str|null}`;
  - `criterios`: lista de `{codigo, ciclo, competencia, texto}`. `ciclo` es uno de los `ciclos` del área y `competencia` es el código de una competencia;
  - `contenidos`: lista de `{ciclo, bloque, subbloque: str|null, nivel: int, texto, pagina: int}`. Los contenidos no llevan código; `nivel` es el anidamiento de las viñetas (1 para «-», 2 para «x», 3 para «») y `pagina` es la página del PDF, para que la revisión pueda localizarlos.
  - Los criterios llevan también `pagina`.
- `recuentos-esperados.json`: `{<area>: {competencias: int, criterios: {<ciclo>: int}, contenidos: {<ciclo>: int}}}`.

**Pasos:**
- [ ] **Escribir las pruebas de integridad**, que fallan porque aún no hay JSON. Todas en `test_datos.py`:
  - `test_codigos_unicos`: no hay ningún código repetido en todo el archivo.
  - `test_formato_codigos`: las competencias cumplen `^(CN|CS|EA|EF|LCL|ING|MAT|VCE|SLE|TYR)\.CE[0-9]+$`, los criterios `^(…)\.[0-9]+\.[0-9]+\.(C1|C2|C3|E)$` (mismos prefijos) y los descriptores `^(CCL|CP|STEM|CD|CPSAA|CC|CE|CCEC)[0-9]$`.
  - `test_areas_y_ciclos`: están exactamente las diez áreas de las restricciones globales, cada una con sus ciclos (VCE solo `C3`; SLE y TYR solo `E`).
  - `test_criterio_apunta_a_competencia`: la `competencia` de cada criterio existe en su área, y el `n` del código del criterio es el `numero` de esa competencia.
  - `test_descriptores_existen`: cada descriptor citado en una competencia está en `descriptores`.
  - `test_equivalencias_cs_cn`: CS.CE1…CE5 → `CMNSC.CE5`…`CE9`; CN.CE1…CE6 → `CMNSC.CE1`…`CE6`; el resto de áreas tiene `null` o la misma numeración del RD.
  - `test_recuentos`: para cada área de `recuentos-esperados.json`, los recuentos coinciden exactamente. Esto detecta los elementos que se pierden al extraer.
  - `test_textos_limpios`: ningún texto queda vacío ni termina en guion de corte de palabra, y ninguno contiene `BOCM-20220718-1`, `BOLETÍN OFICIAL DE LA COMUNIDAD DE MADRID`, `B.O.C.M. Núm.`, `Pág. ` ni `DE JULIO DE 2022`; tampoco quedan los glifos de viñeta `x ` o `` al principio.
  - `test_fuentes`: cada fuente tiene URL `https` y un sha256 de 64 caracteres hexadecimales.
- [ ] Ejecutar `python3 -m venv .venv-dev && .venv-dev/bin/pip install -r requirements-dev.txt && .venv-dev/bin/python -m unittest discover -s test/curriculo -p 'test_datos.py'`. Resultado esperado: FALLA porque no encuentra `files/curriculo/primaria-madrid.json`.
- [ ] **Implementar** `herramientas/extraer-curriculo.py --fuentes cache/fuentes --salida files/curriculo/primaria-madrid.json`:
  - descarga en `cache/fuentes/` las cuatro normas que falten y verifica el sha256 de la nota del spike; si no coincide, aborta;
  - Anexo II: con `pdfplumber.extract_tables()`, página a página. Las filas cuya primera columna está vacía continúan la fila anterior, aunque esté en otra página. Los criterios se separan por su numeración `n.m` (con o sin punto final). Para saber el ciclo, se sigue la última cabecera «PRIMER/SEGUNDO/TERCER CICLO» vista dentro del área; si el área no tiene cabeceras de ciclo, se le asigna su ciclo único;
  - Anexo I: descriptores, sacados del texto por el patrón `^(CCL|CP|STEM|CD|CPSAA|CC|CE|CCEC)[0-9]\.`;
  - RD 157/2022: enlace competencia → descriptores, a partir de la frase «se conecta con los siguientes descriptores del Perfil de salida». Para CN y CS se aplica `equivalencia_rd`. En TYR y SLE, el campo `descriptores` queda vacío salvo que el propio decreto los enumere;
  - los `alias` de cada área se escriben a mano en el extractor (ver los puntos de revisión 5).
- [ ] Generar el JSON. Anotar en `recuentos-esperados.json` los recuentos de la **revisión independiente** (paso siguiente), no los del extractor.
- [ ] **Revisión independiente:** la hace otro LLM con el prompt de `docs/superpowers/prompts/revision-curriculo.md`, en una sesión por área, con el PDF del decreto (y el del RD para los descriptores) y el JSON. Hay que corregir el extractor, nunca el JSON a mano, hasta que cada área tenga el veredicto «APROBADA». El informe de cada área se guarda en `docs/superpowers/revisiones/curriculo-<AREA>.md`.
- [ ] Ejecutar las pruebas de la tarea 1. Resultado esperado: PASAN.
- [ ] Commit: `feat(curriculo): currículo de Primaria de Madrid en JSON con pruebas de integridad`.

---

### Tarea 2: CLI `curriculo` — áreas, consulta y búsqueda

**Archivos:**
- Crear: `files/curriculo/curriculo.py`.
- Crear: `test/curriculo/test_cli.py`.

**Interfaces:**
- Usa: el esquema de la tarea 1. Lee el JSON de la ruta de `$CURRICULO_JSON` o, si no está definida, de `/usr/share/mint-docente/curriculo/primaria-madrid.json`. Las pruebas apuntan al JSON del repo.
- Produce (lo usan la tarea 3 y la skill):
  - `normalizar(texto: str) -> str`: pasa a minúsculas y quita tildes y diéresis (NFKD).
  - `resolver_area(nombre: str) -> list[dict]`: encuentra el área por código, nombre o alias, ya normalizados. Devuelve una lista vacía si no hay ninguna y varias si es ambiguo.
  - `ciclo_para(area: dict, curso: int|None, ciclo: str|None) -> str`:
    - con curso, 1–2 → `C1`, 3–4 → `C2`, 5–6 → `C3`;
    - en las áreas con ciclo `E`, siempre `E`;
    - lanza `ValueError` con un mensaje en español si el curso está fuera de 1–6 o si el ciclo no es de esa área (en VCE: «Valores Cívicos y Éticos solo se imparte en 5º»).
  - `indice_codigos() -> set[str]`: todos los códigos válidos (competencias, criterios y descriptores).
  - Subórdenes, todas con salida en Markdown:
    - `curriculo areas`: tabla de código, nombre, ciclos y alias.
    - `curriculo ver <área> [--ciclo C1|C2|C3|E | --curso N] [--tipo competencias|criterios|contenidos|relaciones|todo]`. Por defecto el tipo es `todo`. `relaciones` muestra, por cada competencia, sus descriptores, su `equivalencia_rd` y sus criterios del ciclo. Los contenidos se muestran agrupados por bloque y sub-bloque, con su sangría.
    - `curriculo buscar <texto> [--area A] [--ciclo … | --curso N]`: busca en criterios, competencias y contenidos. Devuelve `código — texto` o, para los contenidos, `ÁREA ciclo · bloque › sub-bloque — texto`.
    - `curriculo descriptores [<competencia clave>]`.
    - `curriculo fuentes`: normas, fechas y URL.

**Pasos:**
- [ ] **Escribir las pruebas**, que fallan, en `test_cli.py`. Ejecutan la CLI como subproceso con `CURRICULO_JSON` apuntando al JSON del repo:
  - `test_areas_lista_todas`: `areas` sale con 0 y aparecen los diez códigos de área.
  - `test_ver_criterios_ciclo`: `ver MAT --ciclo C2 --tipo criterios` muestra todos los criterios `MAT.*.C2` del JSON y ninguno `.C1` ni `.C3`.
  - `test_curso_a_ciclo`:
    - `ver MAT --curso 3` y `ver MAT --ciclo C2` dan la misma salida;
    - `--curso 7` y `--curso 0` salen con 2 y un mensaje que contiene «1 a 6»;
    - `ver TYR --curso 2` muestra criterios `.E`;
    - `ver VCE --curso 3` sale con 2 y el mensaje contiene «5º».
  - `test_alias`: cada nombre resuelve a su área:
    - «mates», «Matemáticas» y «MAT» → MAT;
    - «ingles» e «Inglés» → ING;
    - «lengua» → LCL;
    - «sociales» → CS;
    - «naturales» → CN;
    - «educación física» → EF;
    - «plástica» y «música» → EA;
    - «robótica» → TYR;
    - «francés» → SLE;
    - «valores» → VCE.
  - `test_alias_ambiguo`: «ciencias» sale con 2 y enumera CN y CS; «idioma» enumera ING y SLE.
  - `test_buscar_sin_tildes`: `buscar division --area MAT` y `buscar DIVISIÓN --area MAT` devuelven las mismas líneas, y al menos una; `buscar lectura --area lengua` devuelve al menos una.
  - `test_buscar_contenidos`: `buscar vertebrados --area naturales --curso 1` devuelve al menos un contenido del bloque «A. Cultura científica».
  - `test_buscar_sin_resultados`: un texto que no está sale con 1 y el mensaje «No hay resultados».
  - `test_relaciones`: en `ver CS --ciclo C1 --tipo relaciones`, cada competencia de CS aparece con su `equivalencia_rd` y al menos un criterio `.C1`.
  - `test_fuentes`: la salida contiene cada URL de `fuentes`.
- [ ] Ejecutar `.venv-dev/bin/python -m unittest test.curriculo.test_cli`. Resultado esperado: FALLA porque no existe `files/curriculo/curriculo.py`.
- [ ] **Implementar** `curriculo.py` con `argparse` y solo la biblioteca estándar. Para buscar, se normaliza el texto y se comprueba si es una subcadena; no hay búsqueda difusa (YAGNI).
- [ ] Ejecutar las pruebas de las tareas 1 y 2. Resultado esperado: PASAN.
- [ ] Commit: `feat(curriculo): CLI curriculo para consultar y buscar el currículo`.

---

### Tarea 3: `curriculo comprobar` — los códigos citados existen

**Archivos:**
- Modificar: `files/curriculo/curriculo.py`, para añadir la suborden.
- Crear: `test/curriculo/test_comprobar.py`.
- Crear: `test/curriculo/ejemplos/`. Los .docx, .xlsx y .md de prueba se generan en la propia prueba, en un directorio temporal, no se versionan.

**Interfaces:**
- Usa: `indice_codigos()` de la tarea 2.
- Produce:
  - `curriculo comprobar <archivo>...`: lee .md, .txt, .docx, .odt, .pdf y .xlsx (los formatos que no son de texto plano, con `markitdown`).
    - Extrae todo lo que **parezca** un código del currículo con una expresión regular derivada del formato de las restricciones globales. La regla es amplia a propósito: también captura variantes mal escritas como `MAT.3.1` sin ciclo, `CE3` sin área o `C.E. 3`.
    - Si todo existe, imprime «Todos los códigos (N) existen en el currículo de Madrid» y sale con 0.
    - Si no, lista cada código desconocido con su archivo y sale con 1. Para cada uno sugiere:
      - los códigos existentes con el mismo prefijo y la misma numeración `n.m` (en todos los ciclos), si existen;
      - si es una competencia escrita con la numeración del RD (`CMNSC.CE7`, o `CE7` en un documento de Ciencias Sociales), la equivalencia de Madrid que dé `equivalencia_rd`.
  - `extraer_codigos(texto: str) -> list[str]`.

**Pasos:**
- [ ] **Escribir las pruebas**, que fallan:
  - `test_md_valido`: un .md que cita tres códigos reales sale con 0 y el mensaje dice «(3)».
  - `test_codigo_inventado`: `MAT.9.9.C2` sale con 1 y aparece en la salida.
  - `test_codigo_estatal_sin_ciclo`: `MAT.3.1`, sin sufijo de ciclo (como lo escriben el currículo estatal y otras comunidades), se detecta como desconocido y se sugieren exactamente los `MAT.3.1.C*` que existan en el JSON.
  - `test_numeracion_rd`: `CMNSC.CE7` se detecta como desconocido y se sugiere `CS.CE3`.
  - `test_docx_en_tabla`: un .docx, creado con python-docx, con un código real y uno inventado **dentro de una tabla** sale con 1 y lista solo el inventado.
  - `test_xlsx`: un .xlsx, creado con openpyxl, con códigos en celdas, se comprueba igual.
  - `test_varios_archivos`: con dos archivos, los códigos desconocidos se atribuyen al archivo correcto.
  - `test_archivo_inexistente`: sale con 2 y nombra el archivo.
  - `test_sin_codigos`: un documento sin códigos sale con 0 y el mensaje «No se ha citado ningún código del currículo».
- [ ] Ejecutar `.venv-dev/bin/python -m unittest test.curriculo.test_comprobar`. Resultado esperado: FALLA porque no existe la suborden `comprobar`.
- [ ] **Implementar.** `markitdown` se importa solo dentro de `comprobar`, así que las demás subórdenes siguen usando únicamente la biblioteca estándar.
- [ ] Ejecutar las pruebas de las tareas 1 a 3. Resultado esperado: PASAN.
- [ ] Commit: `feat(curriculo): comprobar que los códigos citados existen`.

---

### Tarea 4: Integración en la imagen y en el build

**Archivos:**
- Modificar: `provision/20-opencode.sh`, añadiendo un bloque «Currículo de Primaria (Madrid)» después de las skills.
- Modificar: `test/verificar.sh`.
- Modificar: `build.sh`, para que ejecute las pruebas del anfitrión antes de Packer.
- Modificar: `README.md` (secciones «Qué incluye», «Actualizar versiones» y «Estructura»).

**Interfaces:**
- Usa: `files/curriculo/primaria-madrid.json` y `curriculo.py`.
- Produce, en la VM:
  - `/usr/share/mint-docente/curriculo/primaria-madrid.json`;
  - `/usr/local/bin/curriculo`, un envoltorio que ejecuta `curriculo.py` con `/opt/docente-py/bin/python`;
  - `/usr/share/mint-docente/curriculo/<Nombre del área>.md`, una vista legible por área (resultado de `curriculo ver <área> --tipo todo`, con todos los ciclos) para que el docente y el agente puedan abrirla directamente.

**Pasos:**
- [ ] **Escribir las pruebas en `verificar.sh`**, que fallan en un build sin la integración:
  - «CLI curriculo»: `curriculo areas` sale con 0.
  - «Vistas del currículo»: hay un `.md` por cada área del JSON.
  - «Búsqueda sin tildes»: `curriculo buscar division --area mates` sale con 0.
  - «Comprobación de códigos»: con un .docx creado con pandoc que contiene `MAT.9.9.C2`, `curriculo comprobar` sale con 1.
- [ ] **Implementar** el bloque de `20-opencode.sh`: instalar el JSON y `curriculo.py` en `$SHARE/curriculo/`, crear el envoltorio y generar las vistas .md.
- [ ] En `build.sh`, antes de `cd packer`: si existe `.venv-dev/`, ejecutar con él todas las pruebas de `test/curriculo/`; si no existe, ejecutar solo `test_datos.py` y `test_cli.py` con `python3 -I`. En cualquier caso, abortar el build si fallan.
- [ ] Ejecutar en el anfitrión `.venv-dev/bin/python -m unittest discover -s test/curriculo`. Resultado esperado: PASAN todas.
- [ ] Ejecutar `./build.sh`. Resultado esperado: «== Todo correcto», con las cuatro pruebas nuevas en verde.
- [ ] Actualizar el README: fila «Currículo» en «Qué incluye»; en «Actualizar versiones», cómo regenerar el JSON (extractor, recuentos y revisión independiente con el prompt); y las rutas nuevas en «Estructura».
- [ ] Commit: `feat(curriculo): instalar currículo y CLI en la imagen`.

---

### Tarea 5: Instrucciones del agente y skill `curriculo-madrid`

**Archivos:**
- Crear: `files/opencode/skills/curriculo-madrid/SKILL.md`.
- Modificar: `files/opencode/AGENTS.md`, en «Cómo trabajar», «Herramientas disponibles» y «Límites».
- Modificar: `files/opencode/skills/temario/SKILL.md`, en «Antes de empezar».
- Modificar: `test/verificar.sh`, donde «5 skills» pasa a «6 skills».
- Modificar: `files/GUIA-DOCENTE.md`, con un párrafo «Currículo de Madrid».

**Contenido que fija este plan:**
- **AGENTS.md:**
  - Si la etapa no está clara, preguntarla junto con las demás dudas (Primaria o Secundaria).
  - En **Primaria en Madrid**, el currículo de referencia es el de `/usr/share/mint-docente/curriculo/`, que se consulta con `curriculo`.
  - En Madrid se llaman **contenidos**, no «saberes básicos». No tienen código: se citan por bloque y sub-bloque con el texto literal que devuelva `curriculo`.
  - **Regla:** no citar ninguna competencia, criterio ni descriptor que no devuelva `curriculo`. Antes de entregar, ejecutar `curriculo comprobar` sobre el documento final y corregir todo lo que salga como desconocido.
  - Si el docente aporta criterios con otra numeración (por ejemplo, de su libro de texto o del currículo estatal), proponerle la equivalencia de Madrid y no mezclar numeraciones.
  - Programa bilingüe (Decreto 59/2024): Matemáticas, Lengua y los contenidos de Historia de España de Ciencias Sociales no se imparten en inglés.
- **Skill `curriculo-madrid`**, con su descripción en el frontmatter: «Consultar el currículo oficial de Primaria de la Comunidad de Madrid (competencias específicas, criterios de evaluación, contenidos y descriptores del perfil de salida) y comprobar que un documento solo cita códigos que existen». Cuerpo:
  - cuándo usarla;
  - el paso de curso a ciclo (VCE solo en 5º; SLE y TYR, de etapa);
  - qué pide la Orden 130/2023 para cada unidad de programación (art. 6.6), para que el agente sepa qué elementos curriculares debe incluir;
  - el flujo `areas` → `ver --tipo relaciones` → redactar → `comprobar`;
  - cómo citar en el documento: el código y el texto literal, o un resumen marcado como tal;
  - qué decirle al docente si algo no está en el currículo.
- **Skill `temario`:** en el punto 2, si la etapa es Primaria y la comunidad es Madrid, cargar `curriculo-madrid`.

**Pasos:**
- [ ] Cambiar en `verificar.sh` la cuenta de skills a 6 y añadir la prueba «Skill curriculo-madrid» (el archivo existe y tiene `name:` y `description:`). Comprobar que, sin la skill, fallaría.
- [ ] Escribir la skill y los cambios de `AGENTS.md`, `temario` y la guía.
- [ ] Ejecutar `./build.sh`. Resultado esperado: «== Todo correcto».
- [ ] Commit: `feat(curriculo): skill curriculo-madrid y regla de citas en AGENTS.md`.

---

### Tarea 6: Prueba de aceptación con el agente (manual)

Esta tarea no lleva código. Se hace en la OVA recién construida, con una cuenta de IA conectada.

- [ ] Pedirle al agente: «Prepárame la programación de Matemáticas de 3º de Primaria para el primer trimestre». Debe preguntar lo que falte (y no la etapa, porque ya está clara), usar `curriculo` y entregar .docx y .pdf.
- [ ] Ejecutar `curriculo comprobar` sobre el .docx entregado. Resultado esperado: código 0.
- [ ] Pedirle: «Añade el criterio 3.1 del currículo estatal». Debe advertir de la diferencia de numeración y proponer el equivalente de Madrid.
- [ ] Pedir una rúbrica de Lengua de 5º «con los criterios oficiales». `comprobar` debe dar 0 y los textos de los criterios deben coincidir con los de `curriculo ver lengua --curso 5 --tipo criterios`.
- [ ] Anotar los fallos en la sección «Pendiente» del README. Si hay fallos de comportamiento, ajustar la skill o `AGENTS.md` y repetir.

---

## Fuera de este plan (planes siguientes)

- Skills de **programación didáctica de Primaria**, **situaciones de aprendizaje** y **evaluación**: registro .xlsx de criterios → IN/SU/BI/NT/SB e informes a las familias. Se apoyarán en `curriculo` y en la Orden 130/2023 (art. 6: elementos de la programación y de la unidad de programación; escala IN–SB y su equivalencia numérica).
- Atención a la diversidad, documentos de centro, material de aula, programa bilingüe y recursos (fuentes escolares, ARASAAC, clipart, `qrencode`).

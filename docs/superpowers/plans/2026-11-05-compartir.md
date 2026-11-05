# Compartir materiales entre docentes con GitHub — plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Este plan no contiene código.** Por instrucción del usuario, los planes describen decisiones, interfaces y pruebas; el código lo escribe quien ejecuta cada tarea.

**Goal:** Que un docente pueda, si quiere y hablando con el agente, crear un espacio compartido con compañeros, invitarlos, retirarles el acceso, aceptar las invitaciones que reciba y sincronizar los materiales. Todo ello sin ver nunca Git ni GitHub más allá del inicio de sesión.

**Architecture:**
- **Una orden nueva, `compartir`** (Python, solo biblioteca estándar), concentra toda la lógica de GitHub y Git: crear el repositorio, invitar, sincronizar y resolver conflictos sin perder nada. El agente solo llama a `compartir <suborden>`. Así los permisos de OpenCode se pueden fijar por suborden y la lógica delicada (los conflictos) se prueba sin red.
- **Una skill `compartir`** le enseña al agente cuándo y cómo usar la orden, qué decir al docente y qué revisar antes de subir.
- **El inicio de sesión en GitHub es aparte y optativo.** Hay una ventana `mint-docente-github`, con el mismo flujo de código que la de Copilot, que se abre desde el menú o la lanza el agente la primera vez que el docente pide compartir. El primer arranque no cambia.
- **Cada espacio es un repositorio privado de la cuenta personal del docente**, clonado en `Aula/Compartido/<nombre>`. Los colaboradores de una cuenta personal siempre pueden editar: no hay solo lectura.

**Tech Stack:** bash + Packer (imagen), Python 3 de la imagen (biblioteca estándar) para `compartir`, `gh` (GitHub CLI) y `git`, `unittest` para las pruebas en el anfitrión.

**Spec:** decisiones del usuario en la conversación del 2026-11-05:
- se comparte con compañeros, entre **docentes sueltos** (cuentas personales, sin organizaciones);
- los docentes tienen **cuenta de educador** de GitHub;
- es **optativo**: nada en el primer arranque, al contrario que Copilot;
- el agente debe poder **compartir y gestionar el acceso** al repositorio (invitar, quitar, ver quién está);
- lo que se comparte es **sobre todo texto**: aplicaciones pequeñas (HTML, JS, Python), `AGENTS.md` y skills. Los documentos de Office son secundarios.

Consecuencias de lo último en el diseño:
- **Los cambios en texto se mezclan solos.** Si dos docentes tocan partes distintas de un archivo, Git los combina. Solo cuando chocan las mismas líneas se aplica la regla «Dos versiones», y entonces el agente propone una versión combinada.
- **Antes de subir se buscan claves secretas.** Las aplicaciones hechas con agentes suelen llevar claves de API pegadas en el código.
- **Una skill o un `AGENTS.md` de un compañero son instrucciones para tu agente.** Nunca se activan solos: el docente decide qué skills usar (suborden `usar`), se usa una copia fija y las actualizaciones se avisan, no se aplican sin permiso.

Referencia consultada (no se adopta): la skill `publish-to-pages` de `github/awesome-copilot`. Crea un repositorio por archivo, no sabe actualizar y está pensada para usuarios técnicos.

## Global Constraints

- **Rama:** `feat/compartir` en `/datos_nvme/run0/mintDocente`. **Sin push** hasta que lo diga el usuario.
- **Commits** en español, terminando en `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- **Solo repositorios privados de cuenta personal.** Nada de organizaciones, repositorios públicos ni GitHub Pages en esta versión.
- **El docente nunca ve términos de Git.** En la skill, la guía y los mensajes de `compartir`, el vocabulario es: *espacio compartido, compañero, invitar, quitar acceso, traer lo último, guardar en lo compartido, versión*. Nunca *repositorio, commit, push, pull, rama* ni *conflicto*.
- **No se pierde nunca un archivo.** Ante cualquier choque, se conservan las dos versiones.
- **La imagen no lleva credenciales de `gh`**, y el build falla si las lleva.
- **El primer arranque no cambia.** `mint-docente-asistente` y `mint-docente-inicio` siguen haciendo lo mismo para Copilot.
- **`gh` fijado con versión y sha256**, descargado de la release oficial de `cli/cli`, igual que OpenCode en `provision/20-opencode.sh`. El controlador elige la última versión estable el día de la implementación y la pasa en el dispatch.
- **`compartir` usa solo la biblioteca estándar** y se ejecuta con el `python3` del sistema.
- **Variables para pruebas**, con el mismo patrón que `NORMATIVA_*`:
  - `COMPARTIR_GH`: ruta del ejecutable `gh`; por defecto, `gh`.
  - `COMPARTIR_GIT_BASE`: base de las URL de clonado; por defecto, `https://github.com`.
  - `COMPARTIR_RAIZ`: carpeta de los espacios; por defecto, `<Documentos>/Proyecto predeterminado/Compartido`.

## Interfaz de `compartir` (contrato que comparten las tareas)

Las consultas escriben JSON en stdout. Las acciones escriben JSON con el resultado. Los errores se escriben en stderr, en español llano, en una línea pensada para que el agente la traslade al docente.

**Códigos de salida:**
- `0`: correcto.
- `1`: error explicado.
- `3`: sin sesión de GitHub (`SIN_SESION`).
- `4`: sin conexión. El trabajo queda guardado en local.

**Identificar un espacio:** `<espacio>` es el nombre de su carpeta dentro de `COMPARTIR_RAIZ`. El repositorio se deduce del remoto `origin` de esa carpeta.

| Suborden | Qué hace | Salida JSON |
|---|---|---|
| `estado` | Sesión, usuario y espacios locales con cambios sin guardar | `{"usuario": str\|null, "espacios": [{"nombre", "propietario", "pendientes": int}]}` |
| `crear <nombre>` | Crea el repositorio privado `<slug>` con el tema `mint-docente`, lo clona en `RAIZ/<nombre>` y sube un `.gitignore` | `{"espacio", "propietario"}` |
| `invitar <espacio> <usuario>` | Invita a un compañero (solo el propietario) | `{"invitado", "estado": "pendiente"}` |
| `quitar <espacio> <usuario>` | Quita al colaborador o cancela su invitación pendiente | `{"quitado"}` |
| `quien <espacio>` | Lista el propietario, los colaboradores y las invitaciones pendientes | `{"propietario", "colaboradores": [str], "pendientes": [str]}` |
| `invitaciones` | Lista las invitaciones recibidas | `[{"id", "espacio", "de"}]` |
| `aceptar <id>` | Acepta una invitación y clona el espacio en `RAIZ` | `{"espacio"}` |
| `traer [<espacio>]` | Trae lo último de uno o de todos los espacios | `{"espacios": [{"nombre", "nuevos": [str], "mezclados": [str], "dos_versiones": [[str, str]], "skills_con_novedades": [str]}]}` |
| `guardar <espacio> <descripcion>` | Sube los cambios locales, trayendo antes lo último; rechaza la subida si encuentra claves secretas | `{"subidos": [str], "mezclados": [str], "dos_versiones": [[str, str]]}` |
| `usar <espacio> <ruta-skill>` | Copia una skill del espacio a `~/.config/opencode/skills/<nombre>` y anota su origen | `{"skill", "origen", "version"}` |
| `skills` | Lista las skills en uso que vienen de espacios y si tienen versión nueva | `[{"skill", "espacio", "version", "hay_nueva": bool}]` |

**Reglas fijas:**
- **Slug:** minúsculas, sin tildes, y lo que no sea alfanumérico se convierte en `-`. Por ejemplo, «3º ESO Biología» da `3o-eso-biologia`. La carpeta conserva el nombre original.
- **`.gitignore` inicial:** `.~lock.*#`, `~$*`, `*.tmp`, `.DS_Store`, `Thumbs.db`, `.env`, `.env.*`, `node_modules/`, `__pycache__/`, `.venv/`.
- **Claves secretas:** antes de subir, `guardar` revisa los archivos de texto añadidos o cambiados. Si encuentra alguno de estos patrones, termina con código `1` sin subir nada y nombra `archivo:línea`, sin mostrar la clave:
  - `sk-[A-Za-z0-9_-]{20,}` (OpenAI y Anthropic, `sk-ant-…`);
  - `gh[pousr]_[A-Za-z0-9]{30,}` y `github_pat_`;
  - `AIza[0-9A-Za-z_-]{35}` (Google);
  - `-----BEGIN [A-Z ]*PRIVATE KEY-----`;
  - asignaciones con `api_key`, `apikey`, `token`, `secret` o `password` a una cadena literal de 16 caracteres o más.
- **Skills compartidas** (`usar`):
  - Una skill es una carpeta del espacio con un `SKILL.md` que tiene frontmatter con `name` y `description`; si no lo tiene, error.
  - Se **copia** en `~/.config/opencode/skills/<name>`, nunca se enlaza: lo que suba un compañero no cambia lo que hace tu agente sin que lo sepas.
  - Junto a la copia se escribe `.origen` con el espacio, la ruta y el commit copiado.
  - Si ya existe una skill con ese `name` que no viene de este espacio, error con el mensaje «Ya usas otra skill llamada …».
  - `traer` compara cada `.origen` con el espacio y rellena `skills_con_novedades`.
  - Volver a ejecutar `usar` actualiza la copia.
- **`AGENTS.md` compartido:** no tiene suborden. La skill indica al agente que lo trate como una propuesta: enseña al docente qué reglas añadiría y, si las acepta, las incorpora al `AGENTS.md` de la carpeta de trabajo.
- **Identidad de Git**, en la configuración local de cada clon, al crear o aceptar:
  - `user.name`: el nombre del perfil de GitHub, o el usuario si no tiene nombre;
  - `user.email`: `<id>+<usuario>@users.noreply.github.com`.
- **Dos versiones:** cuando el mismo archivo cambia en los dos lados, la versión remota conserva el nombre. La local se guarda como `<base> (versión de <usuario>)<ext>` y ambas se suben.
  - Si un lado borró el archivo y el otro lo modificó, se conserva el modificado.
  - Si ya existe el nombre `(versión de …)`, se añade ` 2`, ` 3`…
- **Tamaño:** ningún archivo de más de 95 MB. `guardar` falla con código `1` antes de subir y nombra el archivo.
- **Rechazo por carrera:** si GitHub rechaza la subida porque alguien subió entre medias, se trae y se reintenta hasta 3 veces.

## Review Focus

1. **Dos docentes cambian el mismo archivo.** Si son líneas distintas, se mezcla solo. Si son las mismas, quedan las dos versiones y el agente propone combinarlas. Nadie pierde trabajo. Pruebas: `test_guardar_texto_se_mezcla` y `test_guardar_mismo_archivo_dos_lados` (tarea 2).
2. **Una aplicación con una clave de API pegada en el código.** No se sube, y se dice en qué archivo y línea está, sin mostrar la clave. Prueba: `test_guardar_rechaza_claves` (tarea 2).
3. **Un compañero cambia una skill que yo ya uso.** Mi agente sigue usando la copia que acepté hasta que yo diga que actualice. Prueba: `test_usar_copia_fija_y_aviso_de_novedad` (tarea 3).
4. **Se va Internet a mitad de `guardar`.** El trabajo queda guardado en local, el código de salida es `4` y el siguiente `guardar` lo sube. Prueba: `test_guardar_sin_conexion_y_reintento` (tarea 2).
5. **Sesión caducada o revocada en github.com.** Debe dar el código `3` y un mensaje que lleve a «Conectar GitHub para compartir», nunca un volcado de error de Python. Prueba: `test_sin_sesion_codigo_3` (tarea 2).

---

### Task 1: `gh` en la imagen y ventana «Conectar GitHub para compartir»

**Files:**
- Modify: `provision/20-opencode.sh`. Bloque nuevo «GitHub CLI», después de OpenCode: descarga del `.deb` con sha256 e instalación.
- Create: `files/escritorio/mint-docente-ventana-codigo`. Función compartida que muestra la ventana del código con zenity; se extrae de `mint-docente-asistente`.
- Modify: `files/escritorio/mint-docente-asistente`. `copilot()` pasa a usar la función compartida, con el mismo comportamiento.
- Create: `files/escritorio/mint-docente-github`.
- Create: `files/escritorio/conectar-github.desktop`.
- Modify: `provision/30-escritorio.sh`. Instala los dos scripts nuevos y el lanzador en el menú. **No** pone icono en el escritorio.
- Modify: `provision/90-limpieza.sh`. Borra `~/.config/gh` y la entrada de `gh` en el llavero, si la hay.
- Modify: `test/verificar.sh`.

**Interfaces:**
- Produces:
  - `mint-docente-github` (en `/usr/local/bin`): sin argumentos. Termina con `0` si la sesión ha quedado iniciada y con `1` si se cancela o falla.
  - `mostrar_codigo <titulo> <codigo> <pid_a_esperar>`: definida en `mint-docente-ventana-codigo`, que se carga con `.` desde `/usr/local/lib/mint-docente/`. Devuelve `0` si el proceso terminó y `1` si el docente canceló.

- [ ] **Step 1: Comprobar a mano la salida de `gh auth login`** (antes de escribir nada)

  En una máquina con la versión de `gh` elegida, ejecutar `gh auth login --hostname github.com --git-protocol https --web --scopes repo` con la variable `BROWSER` apuntando a un ejecutable que no haga nada y con la entrada estándar alimentada con un salto de línea.

  Anotar en el commit:
  - la línea exacta que contiene el código, que se espera con la forma `one-time code: XXXX-XXXX`;
  - que `gh auth status` termina con `0` después de autorizar.

  Si el formato difiere, ajustar la expresión que extrae el código. Es la misma que usa `copilot()`: `[A-Z0-9]{4}-[A-Z0-9]{4}`.

- [ ] **Step 2: Añadir las pruebas a `test/verificar.sh`**

  Pruebas nuevas:
  - «GitHub CLI»: `gh --version` termina con `0`.
  - «Sin credenciales de GitHub»: `gh auth status` termina con un código distinto de `0`, y `~/.config/gh/hosts.yml` no existe.
  - «Lanzador Conectar GitHub»: existe `/usr/share/applications/conectar-github.desktop`.
  - «Sin icono de GitHub en el escritorio»: no existe `$(xdg-user-dir DESKTOP)/conectar-github.desktop`.

- [ ] **Step 3: Implementar la extracción de `mostrar_codigo` y el script `mint-docente-github`**

  - `mostrar_codigo` encapsula, sin cambios de comportamiento, lo que hoy hace `copilot()`:
    - copiar el código al portapapeles;
    - abrir Firefox en `https://github.com/login/device`;
    - mantener la ventana por encima con wmctrl;
    - mostrar la ventana zenity, que se cierra sola cuando termina el proceso.
  - `mint-docente-github`:
    1. Lanza `gh auth login` como en el paso 1 y muestra el código con el título «Mint Docente — Compartir con compañeros».
    2. Al terminar bien, ejecuta `gh auth setup-git` y muestra «✔ Ya puedes compartir materiales. Pídeselo al asistente.».
    3. Si no hay red, muestra el mismo mensaje de error que `copilot()`.
  - Lanzador:
    - `Name=Conectar GitHub para compartir`;
    - `Comment=Compartir materiales con otros docentes (opcional)`;
    - mismas categorías que `cambiar-cuenta-ia.desktop`.
  - En `copilot()` y en `mint-docente-github`, el texto de la ventana dice «inicia sesión en GitHub con tu cuenta de docente».

- [ ] **Step 4: Probar a mano en la VM** (construida con `VER=1 ./build.sh` o con una VM existente actualizada)

  - Abrir «Cambiar cuenta de IA» → GitHub Copilot: funciona igual que antes.
  - Abrir «Conectar GitHub para compartir» desde el menú, autorizar, y comprobar que `gh auth status` termina con `0` y `git ls-remote` funciona con un repositorio privado propio.
  - Cancelar la ventana: no queda ningún proceso `gh` vivo.

- [ ] **Step 5: Commit**: `feat: conexión optativa con GitHub para compartir (gh y ventana de código)`

### Task 2: `compartir`: sesión, crear, traer y guardar sin perder nada

**Files:**
- Create: `files/compartir/compartir`. Ejecutable Python con shebang `/usr/bin/python3`.
- Create: `test/test_compartir.py` (`unittest`) y `test/falso-gh`. El falso `gh` registra cada llamada en `$FALSO_GH_LOG` y responde según `$FALSO_GH_DIR`. Con `repo create`, crea un repositorio bare en `$COMPARTIR_GIT_BASE/<usuario>/<slug>.git`.
- Modify: `provision/20-opencode.sh`. Instala `files/compartir/compartir` en `/usr/local/bin/compartir`.

**Interfaces:**
- Consumes: la sesión de `gh` que deja la tarea 1. En las pruebas, `falso-gh`.
- Produces: las subórdenes `estado`, `crear`, `traer` y `guardar`, con la salida y los códigos de la sección «Interfaz de `compartir`». También las funciones internas que reutiliza la tarea 3:
  - `gh_json(args: list[str]) -> object`: lanza `SinSesion`, `SinConexion` o `ErrorGitHub(mensaje)`.
  - `espacio(nombre: str) -> Espacio(ruta, propietario, repo)`: lanza `ErrorGitHub` si la carpeta no existe o no es un espacio.
  - `configurar_identidad(ruta)`.

- [ ] **Step 1: Escribir las pruebas que fallan** en `test/test_compartir.py`

  Cada prueba trabaja en un directorio temporal. Dos «docentes» son dos `COMPARTIR_RAIZ` distintas sobre el mismo `COMPARTIR_GIT_BASE` (repositorios bare locales), con usuarios `ana` y `luis` en el falso `gh`.

  - `test_sin_sesion_codigo_3`:
    - el falso `gh auth status` falla;
    - `compartir estado` termina con `3`;
    - stderr contiene «Conectar GitHub para compartir»;
    - stderr no contiene `Traceback`.
  - `test_crear_slug_y_carpeta`: `crear "3º ESO Biología"` deja:
    - la carpeta `RAIZ/3º ESO Biología`;
    - la llamada `repo create 3o-eso-biologia --private`, con el tema `mint-docente`;
    - un `.gitignore` que contiene `.~lock.*#`;
    - un `user.email` local que acaba en `@users.noreply.github.com`.
  - `test_crear_nombre_ya_existe`:
    - la segunda llamada a `crear` con el mismo nombre termina con `1`;
    - el mensaje contiene «Ya tienes un espacio llamado»;
    - el falso `gh` no recibe una segunda llamada a `repo create`.
  - `test_guardar_y_traer`:
    - `ana` crea un archivo `ficha.docx` y lo guarda;
    - `luis`, con el espacio clonado, ejecuta `traer`;
    - la lista `nuevos` contiene `ficha.docx` y el archivo es idéntico byte a byte.
  - `test_guardar_mismo_archivo_dos_lados`:
    - `ana` y `luis` modifican `ficha.docx` con contenidos distintos; guarda primero `luis` y después `ana`;
    - en ambos lados, tras `traer`, quedan `ficha.docx` con el contenido de `luis` y `ficha (versión de ana).docx` con el de `ana`;
    - `dos_versiones` es `[["ficha.docx", "ficha (versión de ana).docx"]]`.
  - `test_guardar_texto_se_mezcla`:
    - `ana` cambia la primera línea de `app/index.html` y `luis`, la última;
    - después de que guarden los dos, ambos tienen un único `index.html` con los dos cambios;
    - `mezclados` contiene `app/index.html`;
    - `dos_versiones` está vacío.
  - `test_guardar_rechaza_claves`:
    - un `app/script.js` con `const API_KEY = "sk-ant-` seguido de 40 caracteres hace que `guardar` termine con `1`;
    - stderr contiene `app/script.js:` y la línea, pero no la clave;
    - el remoto no cambia;
    - un `.env` con la misma clave no provoca error, porque está en `.gitignore` y no se sube.
  - `test_borrado_frente_a_modificacion`: `ana` borra `a.docx` y `luis` lo modifica. Tras guardar los dos, `a.docx` existe con el contenido de `luis`.
  - `test_guardar_ignora_bloqueos`:
    - existe `.~lock.ficha.docx#`;
    - `guardar` termina con `0`;
    - `subidos` no contiene el archivo de bloqueo.
  - `test_guardar_archivo_grande`:
    - un archivo de 96 MB (disperso, creado con `truncate`) hace que `guardar` termine con `1` y nombre ese archivo;
    - el remoto no cambia.
  - `test_guardar_sin_conexion_y_reintento`:
    - con `COMPARTIR_GIT_BASE` apuntando a una ruta inexistente, `guardar` termina con `4`;
    - stderr contiene «se subirá la próxima vez»;
    - al restaurar la ruta, otro `guardar` sube el cambio anterior, y `subidos` lo incluye.
  - `test_estado_pendientes`: con un archivo nuevo sin guardar, `estado` da `pendientes: 1` para ese espacio.

- [ ] **Step 2: Ejecutar y ver que fallan.** `python3 -m unittest test/test_compartir.py -v`. Todas fallan porque `compartir` no existe.

- [ ] **Step 3: Implementar `compartir`** con `estado`, `crear`, `traer` y `guardar`, según el contrato.

  **Algoritmo de `guardar`:**
  1. `git add -A`. Rechazar archivos de más de 95 MB y archivos de texto con claves secretas (ver «Reglas fijas»).
  2. Commit con la descripción del docente.
  3. `fetch`.
  4. Si hay cambios remotos, `merge` sin editor. Los archivos de texto que Git mezcle sin conflicto van a `mezclados`.
  5. Si el merge deja archivos en conflicto (texto con las mismas líneas tocadas, o binarios), resolver cada uno con la regla «Dos versiones». Nunca se dejan marcas `<<<<<<<` en un archivo:
     - la versión de `MERGE_HEAD` se queda con el nombre;
     - la de `HEAD` se escribe con el nombre `(versión de <usuario>)`;
     - en borrado frente a modificación, se queda la modificada.
  6. Commit del merge.
  7. `push`. Si se rechaza, volver al paso 3, como máximo 3 veces.
  8. Cualquier fallo de red en `fetch` o `push` → código `4`, con el commit local intacto.

  **`traer`** usa los pasos 3 a 6. Si hay cambios locales sin guardar, los incluye primero en un commit «Cambios locales antes de traer».

  **Mensajes:** todos los de stderr salen de una tabla única al principio del archivo, para que la skill y la guía usen las mismas palabras.

- [ ] **Step 4: Ejecutar y ver que pasan.** `python3 -m unittest test/test_compartir.py -v`: todo en verde.

- [ ] **Step 5: Commit**: `feat: orden compartir para sincronizar espacios sin perder versiones`

### Task 3: `compartir`: invitar, quitar, ver quién está, aceptar invitaciones y usar skills

**Files:**
- Modify: `files/compartir/compartir`.
- Modify: `test/test_compartir.py`, `test/falso-gh`.

**Interfaces:**
- Consumes: `gh_json`, `espacio` y `configurar_identidad` de la tarea 2.
- Produces: las subórdenes `invitar`, `quitar`, `quien`, `invitaciones`, `aceptar`, `usar` y `skills`, según el contrato. Variable de pruebas `COMPARTIR_SKILLS`: carpeta de skills del usuario, por defecto `~/.config/opencode/skills`.

  Llamadas a la API de GitHub:
  - `invitar`: `PUT repos/{o}/{r}/collaborators/{u}`;
  - `quitar`: `DELETE` del colaborador. Si no lo era, busca en `GET repos/{o}/{r}/invitations` y hace `DELETE` de la invitación;
  - `quien`: `GET` de colaboradores e invitaciones;
  - `invitaciones`: `GET user/repository_invitations`, filtrado por el tema `mint-docente`;
  - `aceptar`: `PATCH user/repository_invitations/{id}`, y después clona.

- [ ] **Step 1: Escribir las pruebas que fallan**
  - `test_invitar_y_quien`: `ana invitar <esp> luis` termina con `0`, y `quien` devuelve `pendientes: ["luis"]`.
  - `test_invitar_usuario_inexistente`:
    - el falso `gh` da 404 en `users/nadie`;
    - `invitar` termina con `1` con el mensaje «No hay ningún usuario de GitHub llamado nadie»;
    - no se hace ninguna llamada `PUT`.
  - `test_invitar_con_correo`: `invitar <esp> luis@educa.madrid.org` termina con `1` con el mensaje «Necesito el nombre de usuario de GitHub, no el correo», sin llamadas a la API.
  - `test_invitar_sin_ser_propietario`: si `luis` invita en el espacio de `ana`, termina con `1` con el mensaje «Solo ana puede invitar a este espacio».
  - `test_quitar_colaborador_y_pendiente`: `quitar` elimina tanto a un colaborador aceptado como una invitación pendiente. `quien` ya no muestra a ninguno de los dos.
  - `test_aceptar_clona`:
    - `luis invitaciones` lista la invitación de `ana`;
    - `aceptar <id>` deja la carpeta en la `RAIZ` de `luis`, con su identidad local configurada;
    - una invitación sin el tema `mint-docente` no aparece en la lista.
  - `test_usar_copia_fija_y_aviso_de_novedad`:
    - en el espacio, `ana` sube `skills/rubricas-fp/SKILL.md` con `name: rubricas-fp`;
    - `luis` ejecuta `traer` y `usar <esp> skills/rubricas-fp`, que deja una copia real (no un enlace) en `COMPARTIR_SKILLS/rubricas-fp` con un archivo `.origen`;
    - `ana` modifica la skill y la guarda; `luis` ejecuta `traer`;
    - la copia de `luis` no cambia, y `skills_con_novedades` contiene `rubricas-fp`;
    - `skills` da `hay_nueva: true`;
    - tras otro `usar`, la copia se actualiza y `hay_nueva` pasa a `false`.
  - `test_usar_sin_frontmatter`: con un `SKILL.md` sin `name`, `usar` termina con `1` con el mensaje «no es una skill válida» y no copia nada.
  - `test_usar_nombre_ocupado`: si en `COMPARTIR_SKILLS` ya hay una `rubricas-fp` sin `.origen` o con otro espacio, `usar` termina con `1` con el mensaje «Ya usas otra skill llamada rubricas-fp» y no sobrescribe nada.

- [ ] **Step 0 (antes de las pruebas): Comprobar que OpenCode 2.2.5 carga las skills de `~/.config/opencode/skills/<name>/SKILL.md`.**

  Hacerlo en la VM o con la CLI del anfitrión: crear una skill de prueba ahí y comprobar que el agente la lista.

  Si no la carga, la alternativa es añadir esa ruta a la lista `"skills"` de `/etc/opencode/opencode.jsonc` en la tarea 4. Anotar el resultado en el commit.

- [ ] **Step 2: Ejecutar y ver que fallan.**

- [ ] **Step 3: Implementar.** Antes de cualquier llamada que cambie algo, validar el formato del usuario y comprobar con `GET users/{u}` que existe.

- [ ] **Step 4: Ejecutar todas las pruebas de `test/test_compartir.py`.** Todo en verde.

- [ ] **Step 5: Commit**: `feat: compartir gestiona invitaciones y acceso de compañeros`

### Task 4: skill `compartir`, permisos del agente e instalación

**Files:**
- Create: `files/opencode/skills/compartir/SKILL.md`.
- Modify: `files/opencode/AGENTS.md`.
- Modify: `files/opencode/opencode.jsonc`.
- Modify: `test/verificar.sh`.
- Modify: `README.md`.

**Interfaces:**
- Consumes: `compartir` (tareas 2 y 3) y `mint-docente-github` (tarea 1).

- [ ] **Step 1: Ajustar `test/verificar.sh`**
  - «6 skills» pasa a «7 skills».
  - Prueba nueva «Skill compartir»: el `SKILL.md` tiene `name: compartir` y `description:`.
  - Prueba nueva «Orden compartir sin sesión»: `compartir estado` termina con `3`, porque la imagen no tiene credenciales.
  - Prueba nueva: `python3 -m unittest` sobre `test/test_compartir.py`, copiado a la VM igual que `verificar.sh`.

- [ ] **Step 2: Escribir la skill.** Frontmatter:
  - `name: compartir`;
  - `description`: «Compartir materiales con otros docentes: crear un espacio compartido, invitar o quitar compañeros, ver quién tiene acceso, aceptar invitaciones, traer lo último y guardar los cambios en lo compartido. Úsala cuando el docente quiera compartir o colaborar con compañeros.»

  Lo que debe decir el cuerpo:
  - **Antes de nada**, `compartir estado`. Con código `3`:
    - explica en una frase que hace falta conectar GitHub una vez;
    - ejecuta `mint-docente-github`, que abre la ventana del código;
    - al terminar, repite `compartir estado`.
  - **Correspondencia petición → subórden**, con ejemplos en el lenguaje del docente:
    - «comparte mis fichas de 3º con Ana» → `crear`, copiar los archivos que indique a la carpeta, `guardar`, `invitar`;
    - «¿quién tiene acceso?» → `quien`;
    - «quita a Luis» → `quitar`;
    - «Ana me ha invitado» → `invitaciones` y `aceptar`;
    - «trae lo último» → `traer`;
    - «guarda en lo compartido» → `guardar`.
  - **Pedir el usuario de GitHub del compañero**, no su correo. El propio docente ve el suyo en `compartir estado` y puede dárselo a sus compañeros.
  - **Si piden «solo para ver»:** en GitHub, todo compañero invitado puede editar; no hay forma de evitarlo en una cuenta personal. Se ofrecen dos alternativas: no invitar y mandarle los archivos, o aceptarlo así.
  - **Antes de `guardar`:**
    - revisar con `markitdown` cada archivo nuevo o modificado en busca de datos del alumnado (nombres con notas, DNI, informes, NEE o ACNEE) y de páginas de libros de texto escaneadas;
    - si hay algo, decírselo al docente y no guardar sin su visto bueno explícito.
  - **Al volver a un espacio**, primero `traer`.
  - **Antes de `guardar` una aplicación**, comprobar que no lleva claves ni datos personales. Si `guardar` la rechaza por una clave, explicarlo: «tu aplicación lleva una clave secreta en tal archivo; si la subo, cualquiera del espacio podría usarla a tu costa». Proponer moverla a un archivo `.env`, que no se sube, y que cada compañero ponga la suya.
  - **Antes de `guardar` una skill**, comprobar que su `SKILL.md` tiene frontmatter con `name` y `description`, porque si no, a los compañeros no les funcionará.
  - **Usar una skill de un compañero** («quiero usar la skill de rúbricas de Ana»):
    1. Leer el `SKILL.md` y los archivos de la skill.
    2. Contarle al docente en 2–3 líneas qué hará su agente con ella.
    3. Avisar de cualquier instrucción que suba archivos, use la red, borre cosas, pida claves o contradiga `AGENTS.md`.
    4. Solo con su visto bueno, ejecutar `usar`.
    5. Decirle que cierre y vuelva a abrir OpenCode para que la cargue.
  - **Si `traer` trae `skills_con_novedades`**, avisar al docente («Ana ha cambiado la skill de rúbricas, ¿quieres ver qué ha cambiado?»). Enseñar el cambio con `git diff` sobre la carpeta de la skill, explicado en llano, y ejecutar `usar` solo si el docente acepta.
  - **`AGENTS.md` de un compañero:** no se copia nunca entero. Enseñar al docente qué reglas añadiría a su forma de trabajar y, con su visto bueno, incorporarlas al `AGENTS.md` de su carpeta de trabajo, avisando de las que contradigan las suyas.
  - **Si `mezclados` no está vacío**, decirlo en una línea («he juntado tus cambios con los de Luis en index.html»).
  - **Si `dos_versiones` no está vacío** y los archivos son de texto, comparar las dos versiones y **proponer una combinada**. Escribirla con el nombre original, borrar la copia con confirmación y `guardar`, todo con el visto bueno del docente. Si son binarios, explicarlo así: «Luis también había cambiado la ficha. He guardado las dos: *ficha.docx* (la suya) y *ficha (versión de ana).docx* (la tuya). Dime cuál quieres conservar o si las junto.» Si el docente elige, borrar la que sobra (pide confirmación por `rm`) y `guardar`.
  - **Con código `4`:** «está guardado en tu equipo; se subirá la próxima vez que haya conexión».
  - **Prohibido:**
    - usar `gh` o `git push` directamente;
    - hablar de repositorios, commits o ramas;
    - crear espacios fuera de `Compartido`.

- [ ] **Step 3: Ajustar `AGENTS.md`**
  - En «Límites», la regla «No subas documentos del docente ni datos del alumnado a servicios de Internet» gana una excepción: «salvo con la skill `compartir`, cuando el docente lo pida, y nunca datos del alumnado».
  - Añadir `compartir` a la lista de skills.
  - Excepción a «No crees estructuras de carpetas»: la carpeta `Compartido` la gestiona `compartir`.

- [ ] **Step 4: Ajustar los permisos en `opencode.jsonc`.** La última regla que coincide gana; estas van al final:
  - `gh *` → `deny`;
  - `git push *` → `deny` (ya está);
  - `compartir crear*`, `compartir invitar*`, `compartir quitar*`, `compartir guardar*` y `compartir usar*` → `ask`;
  - el resto de `compartir` y `mint-docente-github` → sin regla, es decir, permitido.

- [ ] **Step 5: README.** Fila nueva en «Qué incluye» (Compartir: `gh`, la orden `compartir` y la skill) y una línea en «Estructura» para `files/compartir/` y `test/test_compartir.py`.

- [ ] **Step 6: Commit**: `feat: skill compartir y permisos del agente para colaborar entre docentes`

### Task 5: guía del docente

**Files:**
- Modify: `files/GUIA-DOCENTE.md`.

- [ ] **Step 1: Escribir la sección «Compartir con compañeros (opcional)»**, después de «Pasar archivos…». Como mucho, una página del PDF. Contenido:
  - **Qué es:** una carpeta *Compartido* dentro de *Aula* que se mantiene igual en tu máquina y en la de tus compañeros.
  - **Conectar una vez:** menú → «Conectar GitHub para compartir», con el mismo código que Copilot. También se abre sola la primera vez que se lo pides al asistente.
  - **Ejemplos de peticiones**, los mismos de la skill, con prioridad a los de aplicaciones y skills:
    - «comparte mi skill de rúbricas con Ana»;
    - «quiero usar la skill de Luis»;
    - «¿hay novedades en las skills que uso?».
  - **Skills de compañeros:** el asistente te explica qué hacen antes de usarlas, y no cambian solas cuando el compañero las modifica.
  - **Claves:** si una aplicación necesita una clave (de una IA, de un servicio), no se comparte; cada uno pone la suya.
  - **Tu usuario de GitHub:** pregúntaselo al asistente («¿cuál es mi usuario de GitHub?») y dáselo a tus compañeros para que te inviten.
  - **Avisos:**
    - quien invitas puede cambiar y borrar archivos del espacio, y siempre se guardan las versiones anteriores;
    - **no compartas datos del alumnado**;
    - si dos personas cambian lo mismo, verás dos archivos y el asistente te preguntará cuál conservar.

- [ ] **Step 2: Comprobar** que el PDF se genera en el anfitrión con la misma orden pandoc de `provision/30-escritorio.sh`, y que la sección cabe en una página.

- [ ] **Step 3: Commit**: `docs: guía para compartir materiales con compañeros`

### Task 6: OVA y prueba de punta a punta con dos cuentas reales

Necesita al usuario: **dos cuentas de GitHub de prueba** y sus inicios de sesión.

- [ ] **Step 1: Construir.** `./build.sh` termina con «Todo correcto», que incluye las pruebas nuevas de las tareas 1 y 4.

- [ ] **Step 2: Recorrido en dos VM importadas de la OVA**, una por cuenta, hablando con el agente en lenguaje de docente:
  1. A: «comparte mi ficha de la célula con <usuario B>». Se abre la ventana de GitHub, se crea el espacio y se pide confirmación para guardar y para invitar.
  2. B: «me han invitado a algo». Acepta, y la ficha aparece en `Aula/Compartido/…`.
  3. A y B modifican la ficha. Guarda B y después A. A ve el aviso de dos versiones con la explicación de la skill.
  4. A comparte una skill. B pide usarla: el agente se la resume, la usa tras su visto bueno y, al reabrir OpenCode, la skill aparece. A la cambia: B recibe el aviso al traer y su skill no cambia hasta que acepta.
  5. A intenta compartir una aplicación con una clave `sk-…` pegada en el código. Se rechaza con la explicación, y el agente propone pasarla a `.env`.
  6. A y B modifican líneas distintas de la misma aplicación: se mezcla sola. Si tocan la misma línea, el agente propone una versión combinada.
  7. A: «¿quién tiene acceso?» y después «quita a B». B pierde el acceso al traer, y el mensaje es comprensible.
  8. A, sin red (cable desconectado en VirtualBox): «guarda en lo compartido». Aparece el aviso de que se subirá más tarde; al reconectar y volver a guardar, se sube.
  9. Primer arranque de una VM nueva: no aparece nada de GitHub salvo el asistente de Copilot.

- [ ] **Step 3: Anotar en el commit de cierre** las discrepancias encontradas y corregidas.

## Fuera de alcance (para un plan siguiente)

- **Web pública con GitHub Pages**, para que las aplicaciones se puedan abrir con un enlace. Con las aplicaciones como contenido principal, es el siguiente paso natural. Con cuenta de educador funciona desde un repositorio privado, pero la web es pública: necesita revisar datos y claves (se puede reutilizar la búsqueda de `guardar`) y una página índice.
- **Organizaciones y permisos finos** (solo lectura, equipos).
- **Recuperar versiones anteriores** («vuelve a la ficha de ayer») y **salir de un espacio** o **borrarlo**. El permiso `repo` no incluye `delete_repo`.

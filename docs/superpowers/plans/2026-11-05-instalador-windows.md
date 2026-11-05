# Instalador para Windows, OVA más ligera y «Mi trabajo» — plan de implementación

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Este plan no contiene código.** Por instrucción del usuario, los planes describen decisiones, interfaces y pruebas; el código lo escribe quien ejecuta cada tarea.

**Goal:** Que un docente instale Mint Docente en Windows con un doble clic y sin permisos de administrador, que la descarga pese ~3,5 GB y que en un equipo del centro que se borra pueda guardar y traer su trabajo con GitHub hablando con el asistente.

**Architecture:** Tres partes independientes que se pueden entregar por separado, en este orden:
- **A — OVA ligera:** un script de aprovisionamiento nuevo, `provision/80-recorte.sh`, quita lo que sobra. Después se reinicia la máquina y `test/verificar.sh` comprueba que todo sigue funcionando.
- **C — «Mi trabajo»:** una suborden nueva, `compartir mi-trabajo`, convierte la carpeta Aula en un repositorio privado del docente y sabe restaurar los espacios y las skills en una máquina nueva. La skill y el AGENTS.md enseñan al agente a usarla.
- **B — Instalador:** un `.exe` en Go compilado desde Linux, con diálogos nativos de Windows, que habla con `VBoxManage`. La lógica vive en paquetes probados con `go test` en Linux; lo específico de Windows va en archivos `_windows.go`.

**Tech Stack:** bash y Packer (imagen); Python 3 con la biblioteca estándar (`compartir`) y `unittest`; Go para Windows/amd64 con `github.com/ncruces/zenity` (diálogos), `github.com/mattn/go-ieproxy` (proxy del sistema), `golang.org/x/sys/windows` y `github.com/go-ole/go-ole` (acceso directo), y `github.com/tc-hib/go-winres` (manifiesto e icono).

**Spec:** `docs/superpowers/specs/2026-11-05-instalador-windows-design.md` (aprobado el 2026-11-05).

## Global Constraints

- Todo lo que ve el docente, en español llano y en ventanas. Nunca una consola ni un mensaje técnico.
- El instalador no pide permisos de administrador: el manifiesto lleva `requestedExecutionLevel` = `asInvoker`, y el ejecutable usa el subsistema GUI (`-H=windowsgui`).
- Nombre del ejecutable: `InstalarMintDocente.exe`, sin firmar.
- Nombre de la máquina en VirtualBox: `Mint Docente`.
- Origen de la descarga: `https://github.com/balejosg/mint-docente/releases/latest/download/`. Las partes y sus sumas se leen de `SHA256SUMS`, cuyo formato no cambia: una línea por parte `*.ova.partN` y una línea para la `.ova` completa.
- Mensajes con texto fijado por el spec: «Este equipo no tiene VirtualBox. Pídeselo al coordinador TIC.» y «Este VirtualBox es antiguo; pide al coordinador TIC la versión 7.x».
- Imagen: `test/verificar.sh` pasa entero. OVA de ~3,5 GB: si pasa de 3,8 GB, se avisa al usuario antes de publicar.
- `compartir`: salida en JSON, errores en una línea de español y los mismos códigos de salida (0, 1, 3 y 4). Las claves nunca se suben. Los datos del alumnado se suben solo con el visto bueno explícito del docente (lo controla la skill, como hoy).
- Las dependencias descargadas en el build se fijan con su versión y su sha256, como ya se hace con OpenCode y gh.

## Review Focus

1. **Proxy del centro:** en una red con proxy configurado en Windows (incluso con PAC), la descarga debe funcionar. Go no usa el proxy del sistema por sí solo → Task B3.
2. **Descarga cortada o disco lleno a mitad:** al volver a abrir el instalador, sigue donde lo dejó y no vuelve a empezar. Si el disco se llena, lo dice, no deja la máquina a medias y no deja basura → Tasks B3 y B7.
3. **Documentos redirigido a una ruta de red (`\\servidor\…`):** Intercambio no debe romper la importación. Si la carpeta compartida no se puede añadir, la máquina arranca igual y se avisa → Task B5.
4. **Usuario de Windows con tildes o espacios** (`C:\Users\José Ángel`): todas las rutas llegan intactas a `VBoxManage` → Task B5.
5. **Reinstalar con la máquina encendida:** se le pide al docente que la cierre. No se borra nada mientras esté en marcha → Task B7.

---

## Parte A — OVA ligera

### Task A1: Recorte de la imagen

**Files:**
- Create: `provision/80-recorte.sh`
- Modify: `packer/mint-docente.pkr.hcl`: un provisioner nuevo después de `30-escritorio.sh`, con `expect_disconnect = true`, que ejecuta `80-recorte.sh` y reinicia. El provisioner de `verificar.sh` lleva `pause_before = "30s"` para que verifique después de ese reinicio.
- Modify: `test/verificar.sh`: pruebas nuevas.
- Modify: `README.md` (tabla «Qué incluye»: ya no hay Thunderbird ni compiladores) y `provision/90-limpieza.sh` si hace falta.

**Decisiones:**
- **Kernel:** se purgan todos los paquetes de kernel (`linux-image-*`, `linux-modules-*` y `linux-headers-*`) de versiones distintas de `uname -r`. Tras el reinicio de 00-base, la versión en marcha es la más nueva. Se **retienen** con `apt-mark hold` los metapaquetes del kernel (`linux-generic-hwe-24.04`, `linux-image-generic-hwe-24.04` y `linux-headers-generic-hwe-24.04`, o los que estén instalados). Se desactiva la actualización automática de Mint (`mintupdate-automation-upgrade.timer` y `unattended-upgrades`). Así no puede llegar un kernel nuevo sin Guest Additions.
- **Compiladores:** con el kernel retenido, se purgan `build-essential`, `dkms` y las cabeceras del kernel en marcha. Antes se comprueba que el módulo `vboxguest` del kernel en marcha está instalado fuera de DKMS o que DKMS ya lo ha instalado en `/lib/modules/$(uname -r)`.
- **Aplicaciones:** se purgan, si están instaladas, `thunderbird*`, `hexchat*`, `transmission-*`, `hypnotix`, `celluloid`, `rhythmbox*`, `warpinator` y `timeshift`.
- **Fuentes:** se purgan `fonts-noto-cjk*`, `fonts-noto-extra` y `fonts-noto-ui-extra`. `fonts-noto-core` se queda.
- **Iconos:** se conserva el tema que usa el docente (se lee de su configuración de xfconf, `/Net/IconThemeName`), su tema de respaldo (`Inherits` en `index.theme`) y `hicolor`. Se purgan los paquetes de los demás temas que haya en `/usr/share/icons`, comprobando cada uno con `dpkg -S`.
- **Firmware, documentación y locales:** un archivo `/etc/dpkg/dpkg.cfg.d/90-mint-docente` con `path-exclude` para `/usr/lib/firmware/*`, `/usr/share/doc/*` (salvo `copyright`), `/usr/share/man/*` y `/usr/share/locale/*`, con `path-include` para `es*`, `ca*`, `gl*`, `eu*` y `en*`. Después se borran los archivos que ya existen bajo esas rutas, con las mismas reglas. Así las actualizaciones no los vuelven a poner.
- **Fondos:** se conserva el fondo configurado del escritorio y se borran los demás de `/usr/share/backgrounds`.
- **Se conservan:** Java (`/usr/lib/jvm`, que LibreOffice usa en algunos asistentes), TeX Live y todo lo que figura en la tabla del README.

- [ ] **Step 1: Pruebas nuevas en `test/verificar.sh`**
  - «Módulo vboxsf»: `lsmod` incluye `vboxsf`.
  - «Portapapeles de VirtualBox»: existe `VBoxClient` y hay un proceso `VBoxClient --clipboard` en la sesión del docente.
  - «Kernel retenido»: `apt-mark showhold` incluye al menos un metapaquete `linux-image-*`.
  - «Un solo kernel»: hay exactamente un `vmlinuz-*` en `/boot`.
  - «Sin Thunderbird»: no existe el comando `thunderbird`.
  - «Fuentes para documentos»: `fc-list` tiene Noto Sans, DejaVu Sans, Carlito y Arial.
  - «Locales en español»: existe `/usr/share/locale/es/LC_MESSAGES` y no existe `/usr/share/locale/ja`.
  - «LibreOffice con Java»: `soffice --headless --convert-to odt` de un `.docx` funciona (ya cubierto) y `java -version` responde.

- [ ] **Step 2: Escribir `provision/80-recorte.sh`** con las decisiones anteriores. Usa `lib.sh` (`log`, `apt_get`) como el resto de scripts. Termina con `apt_get autoremove --purge` y reinicia.

- [ ] **Step 3: Build completo**
  Run: `./build.sh`
  Expected: termina sin fallos de `verificar.sh` y la OVA de `salida/` pesa ≤ 3,8 GB (`stat -c %s`). Si pesa más, se anotan las 15 rutas más grandes (lo que se midió en el spec) y se avisa al usuario.

- [ ] **Step 4: Prueba manual en una VM importada** de la OVA nueva: la carpeta compartida automontada aparece, copiar y pegar funciona en los dos sentidos, la pantalla se ajusta al redimensionar y OpenCode abre.

- [ ] **Step 5: Commit**: `feat: imagen más ligera (sin firmware, Thunderbird ni compiladores; kernel retenido)`.

---

## Parte C — «Mi trabajo» en GitHub

### Task C1: `compartir mi-trabajo estado` y `guardar`

**Files:**
- Modify: `files/compartir/compartir`
- Modify: `test/test_compartir.py`, `test/falso-gh` (si hace falta admitir `repo create` con otra descripción)

**Interfaces (Produces):**
- `compartir mi-trabajo estado` → `{"en_github": bool, "local": bool, "pendientes": int}`. Código 3 sin sesión, 4 sin conexión.
- `compartir mi-trabajo guardar <descripción>` → `{"subidos": [str], "mezclados": [str], "dos_versiones": [str], "no_subidos": [{"archivo": str, "motivo": "grande"}], "espacios": [{"nombre": str, "subidos": [...], "mezclados": [...], "dos_versiones": [...]}]}`.
- Funciones nuevas: `aula() -> Path`, que devuelve `raiz().parent`, y `datos_mi_trabajo() -> Path`, que devuelve `aula() / ".mint-docente"`.
- Repositorio: `mint-docente-mi-trabajo`, privado, con la descripción `Mi trabajo de Mint Docente`. No empieza por `DESCRIPCION_REPO`, así que nunca sale en `invitaciones`.
- Metadatos en `aula()/.mint-docente/`:
  - `espacios.json`: lista de `{"nombre", "propietario", "repo"}` de cada espacio de `Compartido/`;
  - `skills-companeros.json`: lista de `{"skill", "espacio", "ruta", "version"}`, tomada de `skills_en_uso()`;
  - `skills-propias/<name>/…`: copia de cada skill de `skills_dir()` que no tiene `.origen`.

**Decisiones:**
- El `.gitignore` del Aula es `GITIGNORE` más `Compartido/`.
- `guardar`:
  1. Si no existe el repositorio, lo crea (como `cmd_crear`, pero sobre `aula()` con `git init` y el remoto).
  2. Reescribe los metadatos.
  3. Escribe en `.git/info/exclude` los archivos de más de `LIMITE_BYTES` y los devuelve en `no_subidos`, en vez de fallar. `comprobar_salientes` recibe un parámetro `grandes="error"|"excluir"`.
  4. Las claves siguen fallando con el mensaje `clave`.
  5. Usa el mismo bucle de `cmd_guardar` (factorizado en una función `guardar_repo(esp_o_ruta, descripcion) -> dict`).
  6. Después ejecuta `guardar_repo` sobre cada espacio local con `pendientes > 0`.

- [ ] **Step 1: Pruebas que fallan** (clase `MiTrabajo` en `test/test_compartir.py`):
  - `test_estado_sin_repo`: `en_github` es False y `local` es False.
  - `test_guardar_crea_repo_privado`: tras guardar con `notas.md` en el Aula, el gh falso registra un `repo create mint-docente-mi-trabajo --private` y el repositorio bare contiene `notas.md` y `.mint-docente/espacios.json`.
  - `test_guardar_no_sube_compartido`: con un espacio en `Compartido/Rúbricas`, el repositorio de mi-trabajo no contiene rutas bajo `Compartido/`.
  - `test_guardar_excluye_grande`: un archivo de 96 MB sale en `no_subidos` con `motivo` = `grande` y no está en el repositorio; el código de salida es 0.
  - `test_guardar_rechaza_clave`: un archivo con `CLAVE_FALSA` da código 1, el mensaje contiene «clave secreta» y no se sube nada.
  - `test_guardar_skills_propias_y_de_companeros`: una skill propia en `skills/ana/rubricas` aparece en `.mint-docente/skills-propias/rubricas/SKILL.md`. Una skill usada con `compartir usar` aparece en `skills-companeros.json` con su `version`.
  - `test_guardar_tambien_espacios_pendientes`: un cambio en un espacio compartido sale en `espacios[0].subidos`.
  - `test_mi_trabajo_no_es_invitacion`: tras guardar, `compartir invitaciones` de otro usuario no lo lista.
  - `test_ruta_con_tildes`: Aula en `…/José Ángel/Proyecto predeterminado`: guardar funciona.
- [ ] **Step 2: Ejecutarlas**. Run: `python3 -m unittest test/test_compartir.py -v -k MiTrabajo`. Expected: FAIL.
- [ ] **Step 3: Implementar** según las interfaces y decisiones. Actualizar `MENSAJES["uso"]`.
- [ ] **Step 4: Toda la batería**. Run: `python3 -m unittest test/test_compartir.py -v`. Expected: PASS, con las 30 pruebas anteriores incluidas.
- [ ] **Step 5: Commit**: `feat(compartir): guardar todo el Aula en «Mi trabajo»`.

### Task C2: `compartir mi-trabajo traer`

**Files:** los mismos que en C1.

**Interfaces:**
- Consumes: `aula()`, `datos_mi_trabajo()`, los metadatos de C1, `clonar()`, `mezclar()` y `skills_dir()`.
- Produces: `compartir mi-trabajo traer` → `{"traidos": [str], "mezclados": [str], "dos_versiones": [str], "espacios_recuperados": [str], "skills_recuperadas": [str], "avisos": [str]}`. Si no existe el repositorio en GitHub: código 1 con el mensaje nuevo `MENSAJES["sin_mi_trabajo"]` = «Todavía no has guardado tu trabajo en GitHub desde ningún equipo.».

**Decisiones:**
- Si `aula()` no es un repositorio, se hace `git init` con el remoto. Si hay archivos locales, primero se confirman («Cambios locales antes de traer»). Después se mezcla con `--allow-unrelated-histories`, siguiendo la regla de `dos_versiones` de `mezclar()`. Nunca se pierde un archivo local.
- **Espacios:** cada entrada de `espacios.json` que no exista en `Compartido/` se clona con `clonar()`. Si da 404, se añade un aviso con `MENSAJES["sin_acceso_espacio"]` y se sigue.
- **Skills de compañeros:** se restauran **en la versión guardada**. El árbol `version` del espacio se extrae con `git archive`, sin tomar la de HEAD, para mantener la «copia fija». Se escribe el mismo `.origen`. Si la versión no existe o el nombre choca con una skill del sistema o con una ya presente, se añade un aviso y se sigue.
- **Skills propias:** se copian de `skills-propias/` a `skills_dir()` si no existen. Si ya existe una distinta, se añade un aviso y no se pisa.

- [ ] **Step 1: Pruebas que fallan:**
  - `test_traer_en_maquina_nueva`: ana guarda (Aula con `notas.md`, un espacio propio, un espacio aceptado de luis y una skill de luis en uso). Después se borran el Aula y las skills de ana y se ejecuta `traer`. Vuelven `notas.md` y los dos espacios en `Compartido/`, y la skill tiene `.origen` con la misma `version`.
  - `test_traer_skill_en_version_fija`: luis cambia su skill después de que ana guarde; tras `traer`, el `SKILL.md` de ana es el de antes.
  - `test_traer_no_pisa_local`: el Aula local tiene `notas.md` con otro contenido; tras `traer` existen `notas.md` y `notas (versión de ana).md` (el formato que ya usa `mezclar`), y `dos_versiones` lo lista.
  - `test_traer_espacio_sin_acceso`: luis quita a ana de su espacio; `traer` termina con código 0, el aviso menciona el acceso y el resto se recupera.
  - `test_traer_sin_repo`: código 1 y el mensaje `sin_mi_trabajo`.
  - `test_traer_sin_sesion`: código 3.
- [ ] **Step 2: Ejecutarlas**. Expected: FAIL.
- [ ] **Step 3: Implementar.**
- [ ] **Step 4: Toda la batería**. Expected: PASS.
- [ ] **Step 5: Commit**: `feat(compartir): traer «Mi trabajo» a una máquina nueva`.

### Task C3: Desconectar cuentas

**Files:**
- Create: `files/escritorio/mint-docente-desconectar` y `files/escritorio/desconectar-cuentas.desktop` (menú: «Desconectar mis cuentas»)
- Create: `test/probar-desconectar.sh`
- Modify: `provision/30-escritorio.sh` (instalar los dos archivos) y `test/verificar.sh` (prueba «Lanzador Desconectar»)

**Decisiones:**
- Una ventana de zenity pregunta: «¿Desconecto tus cuentas de IA y de GitHub de este equipo? Tendrás que volver a conectarlas la próxima vez.».
- Si el docente acepta: `gh auth logout -h github.com` (si hay sesión) y borrar el archivo de credenciales de OpenCode (`~/.local/share/opencode/auth.json`, la ruta que ya limpia `90-limpieza.sh`; confirmar el nombre en la VM).
- Termina con un aviso de que hay que cerrar y volver a abrir OpenCode.
- Admite `--si` para las pruebas, que se salta la pregunta.

- [ ] **Step 1: `test/probar-desconectar.sh`**, con un HOME temporal, un `gh` falso en el PATH que registra las llamadas y un `auth.json` falso. Tras `mint-docente-desconectar --si`, `auth.json` no existe y el registro tiene `auth logout -h github.com`. Sin `--si` y con zenity falso respondiendo «no», no se toca nada.
- [ ] **Step 2: Ejecutarlo**. Run: `bash test/probar-desconectar.sh`. Expected: FAIL.
- [ ] **Step 3: Implementar e instalar.**
- [ ] **Step 4: Ejecutarlo**. Expected: PASS.
- [ ] **Step 5: Commit**: `feat: desconectar las cuentas de un equipo compartido`.

### Task C4: Skill, AGENTS.md, permisos y guía

**Files:**
- Modify: `files/opencode/skills/compartir/SKILL.md`: sección «Mi trabajo» y filas en la tabla «El docente dice → Órdenes».
- Modify: `files/opencode/AGENTS.md`: reglas de inicio y recordatorio.
- Modify: `files/opencode/opencode.jsonc`: `{ "action": "shell", "resource": "compartir mi-trabajo guardar*", "effect": "ask" }`.
- Modify: `files/GUIA-DOCENTE.md` y `README.md`.

**Contenido que fija el spec:**
- Tabla:
  - «Guarda mi trabajo» → `compartir mi-trabajo guardar "<qué ha hecho hoy>"`, con las mismas revisiones de «Antes de guardar».
  - «Trae mi trabajo» → `compartir mi-trabajo traer`.
  - Si el docente dice que se va: después de guardar, ofrece la ventana `setsid mint-docente-desconectar >/dev/null 2>&1 &`.
- `no_subidos`: el agente nombra los archivos y avisa de que en un equipo del centro se perderán al cerrar la sesión, salvo que los copie a Intercambio.
- AGENTS.md, inicio: si la carpeta de trabajo está vacía (o solo tiene `Compartido/`), ejecuta `compartir mi-trabajo estado`. Con `en_github` = true, ofrece traer el trabajo. Con el código 3, pregunta en una frase si quiere conectar GitHub para traer el trabajo guardado desde otro equipo.
- AGENTS.md, recordatorio: al terminar cada encargo, una línea: «Si trabajas en un equipo del centro, pídeme que guarde tu trabajo antes de irte.». **El agente nunca guarda sin que se lo pidan.**
- La skill dice qué entra en «Mi trabajo» y qué no: no entra el historial de conversaciones.
- Guía: una sección «Si trabajas en un ordenador del centro» con «Guarda mi trabajo», «Trae mi trabajo» y «Desconectar mis cuentas».

- [ ] **Step 1: Pruebas en `test/verificar.sh`**: «Permiso mi-trabajo guardar» (el `opencode.jsonc` instalado contiene `compartir mi-trabajo guardar*` con `ask`) y «Skill compartir conoce mi-trabajo» (grep de `mi-trabajo traer` en la skill instalada).
- [ ] **Step 2: Redactar los cambios.**
- [ ] **Step 3: Prueba de comportamiento en una VM** construida con A y C, con una cuenta de GitHub de pruebas. Pedir «guarda mi trabajo», importar otra copia de la OVA, conectar GitHub y comprobar que el agente ofrece traerlo y que lo recupera todo. Pedir «me voy» y comprobar que ofrece desconectar.
- [ ] **Step 4: Commit**: `feat: el asistente guarda y trae el trabajo del docente`.

---

## Parte B — Instalador para Windows

### Task B1: Spike en Windows (código desechable)

Necesita un **Windows real o virtual con VirtualBox y una cuenta de usuario estándar**. Si quien ejecuta no lo tiene, se detiene y se lo pide al usuario.

- [ ] **Step 1:** Un `.exe` mínimo en Go con el manifiesto `asInvoker`. Muestra un diálogo de zenity, ejecuta `VBoxManage --version` (localizado como en B5) e importa la OVA de la Parte A con `--basefolder` en el perfil del usuario.
- [ ] **Step 2:** Con la cuenta estándar, anotar:
  - qué muestra SmartScreen (con una captura para la guía);
  - si el manifiesto evita la elevación;
  - si la importación y `sharedfolder add --automount` funcionan;
  - si `go-ieproxy` lee el proxy;
  - la versión mínima de VirtualBox que importa la OVA (probar con 7.0 si es posible).
- [ ] **Step 3:** Resultados en `docs/superpowers/notas/AAAA-MM-DD-spike-instalador-windows.md`. Fijar `VBOX_MINIMA` (provisional: `7.0`). Si algo invalida el diseño (la importación exige administrador, por ejemplo), parar y hablarlo con el usuario.
- [ ] **Step 4: Commit** de la nota (no del código del spike).

### Task B2: Módulo, toolchain y sumas

**Files:**
- Create: `instalador/go.mod` (módulo `github.com/balejosg/mint-docente/instalador`), `instalador/construir.sh`, `instalador/internal/sumas/sumas.go` y `sumas_test.go`, y `instalador/winres/winres.json` (manifiesto `asInvoker`, icono y nombre de producto «Mint Docente»).

**Interfaces (Produces):**
- `sumas.Leer(r io.Reader) (sumas.Lista, error)`, donde `Lista{Partes []Archivo; Completa Archivo}` y `Archivo{Nombre string; SHA256 string}`. Las partes van ordenadas por el número N de `.partN`, no alfabéticamente.
- `construir.sh`:
  1. Descarga Go a `cache/go/` si falta, con la versión estable vigente en el momento de implementar y su sha256 de go.dev/dl, fijados en el script.
  2. Ejecuta `go test ./...`.
  3. Embebe los recursos con `go-winres` (versión fijada).
  4. Compila con `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` y `-ldflags "-H=windowsgui -s -w"` → `salida/InstalarMintDocente.exe`.

- [ ] **Step 1: Pruebas:**
  - `TestLeerSHA256SUMSv11`: con el contenido real de `salida/release-v1.1/SHA256SUMS` (copiado como dato de prueba), da 3 partes en orden y la completa `MintDocente-20261105.ova`.
  - `TestOrdenNumerico`: `part10` va después de `part9`.
  - `TestSinCompleta`: error.
  - `TestLineaMalFormada`: error.
- [ ] **Step 2: Ejecutarlas**. Run: `cd instalador && go test ./...`. Expected: FAIL.
- [ ] **Step 3: Implementar.**
- [ ] **Step 4:** Run: `instalador/construir.sh`. Expected: PASS y existe `salida/InstalarMintDocente.exe`. `file` dice «PE32+ executable (GUI) x86-64».
- [ ] **Step 5: Commit**: `feat(instalador): módulo Go, construcción y lectura de SHA256SUMS`.

### Task B3: Descarga reanudable con proxy del sistema

**Files:** `instalador/internal/descarga/descarga.go`, `descarga_test.go`

**Interfaces (Produces):**
- `descarga.Bajar(ctx context.Context, cli *http.Client, url, destino string, sha256 string, progreso func(hechos, total int64)) error`. Escribe en `destino + ".descargando"`; si ese archivo existe, continúa con `Range`. Al terminar comprueba el sha256 y renombra. Si la suma no cuadra, borra el archivo y devuelve `descarga.ErrSuma`.
- `descarga.Cliente() *http.Client`: transporte con el proxy de `go-ieproxy` y reintentos (3, con espera creciente) ante errores de red.

- [ ] **Step 1: Pruebas** con `httptest` y un servidor que corta la conexión a mitad:
  - `TestReanuda`: segunda llamada con cabecera `Range: bytes=<n>-` y archivo final correcto.
  - `TestServidorSinRange`: si responde 200 en vez de 206, empieza de cero sin corromper el archivo.
  - `TestSumaIncorrecta`: devuelve `ErrSuma` y no deja ni `destino` ni `.descargando`.
  - `TestRedireccion`: sigue el 302 de GitHub a otro host.
  - `TestDiscoLleno`: un escritor que falla con `ENOSPC` hace que se devuelva `descarga.ErrSinEspacio`.
  - `TestUsaProxy`: con el proxy configurado (`HTTPS_PROXY` en la prueba), la petición pasa por un `httptest` que hace de proxy.
- [ ] **Steps 2–4:** fallar, implementar, pasar (`go test ./internal/descarga/`).
- [ ] **Step 5: Commit**: `feat(instalador): descarga reanudable con el proxy del sistema`.

### Task B4: Unir partes

**Files:** `instalador/internal/unir/unir.go`, `unir_test.go`

**Interfaces:** `unir.Partes(partes []string, destino string, sha256 string) error`. Concatena en orden, borra cada parte después de añadirla y comprueba la suma de la completa. Si no cuadra: `unir.ErrSuma`, se borra `destino` y las partes ya no están (el flujo vuelve a descargar).

- [ ] **Step 1: Pruebas:**
  - `TestUneEnOrden`: tres partes pequeñas y suma correcta.
  - `TestBorraPartes`.
  - `TestSumaMal`: no queda `destino`.
- [ ] **Steps 2–4:** fallar, implementar, pasar.
- [ ] **Step 5: Commit**: `feat(instalador): unir las partes de la OVA`.

### Task B5: VirtualBox

**Files:** `instalador/internal/vbox/vbox.go`, `localizar_windows.go`, `localizar_otros.go` (para las pruebas) y `vbox_test.go`

**Interfaces (Produces):**
- `type Ejecutor interface { Ejecutar(args ...string) (salida string, err error) }`, con una implementación real sobre `exec.Command(VBoxManage, …)` y con `HideWindow` para que no aparezca una consola.
- `vbox.Localizar() (ruta string, err error)`: busca en el registro `HKLM\SOFTWARE\Oracle\VirtualBox` el valor `InstallDir`, después en `VBOX_MSI_INSTALL_PATH` y después en `%ProgramFiles%\Oracle\VirtualBox`. Si no lo encuentra: `vbox.ErrNoInstalado`.
- `vbox.Version(e Ejecutor) (Version, error)`, que interpreta `7.2.4r170995`, y `Version.Menor(otra Version) bool`.
- `vbox.Estado(e Ejecutor, nombre string) (existe, encendida bool, err error)`, con `list vms` y `list runningvms`.
- `vbox.CarpetaMaquinas(e Ejecutor) (string, error)`: lee «Default machine folder:» de `list systemproperties`.
- `vbox.Importar(e Ejecutor, ova, nombre, carpeta string, memoriaMB, cpus int) error` → `import <ova> --vsys 0 --vmname <nombre> --basefolder <carpeta> --memory <m> --cpus <c>`.
- `vbox.Compartir(e Ejecutor, nombre, ruta string) error` → `sharedfolder add <nombre> --name Intercambio --hostpath <ruta> --automount`.
- `vbox.Arrancar(e Ejecutor, nombre string) error` → `startvm <nombre> --type gui`.
- `vbox.Borrar(e Ejecutor, nombre string) error` → `unregistervm <nombre> --delete`.

- [ ] **Step 1: Pruebas** con un `Ejecutor` falso que registra los argumentos:
  - `TestImportarArgs`: la lista exacta de argumentos, con `carpeta` = `C:\Users\José Ángel\VirtualBox VMs` como **un solo argumento** intacto.
  - `TestVersion`: `7.2.4r170995` > `7.0`, y `6.1.50` es menor que `7.0`.
  - `TestEstado`: con la salida `"Mint Docente" {uuid}`, `existe` es true; con `runningvms` vacío, `encendida` es false.
  - `TestCompartirRutaRed`: con la ruta `\\servidor\docs\Intercambio`, el argumento llega intacto. Que falle o no lo decide el flujo (B7).
  - `TestCarpetaMaquinas`.
- [ ] **Steps 2–4:** fallar, implementar, pasar.
- [ ] **Step 5: Commit**: `feat(instalador): órdenes de VirtualBox`.

### Task B6: Recursos del equipo

**Files:** `instalador/internal/equipo/equipo.go`, `equipo_windows.go`, `equipo_otros.go` y `equipo_test.go`

**Interfaces (Produces):**
- `equipo.Ajuste(memoriaTotalMB int, nucleos int) (memoriaMB, cpus int, lenta bool)`:
  - la memoria es la mitad de la total, redondeada hacia abajo a múltiplos de 512 y limitada a [2048, 4096];
  - los procesadores son la mitad de los lógicos, limitados a [1, 4];
  - `lenta` es true si la memoria total es menor de 6144 MB.
- `equipo.EspacioNecesario(tamOVA int64) int64` = 2 × `tamOVA` + 12 GiB.
- `equipo_windows.go`: `MemoriaTotalMB()` (`GlobalMemoryStatusEx`), `EspacioLibre(ruta)` (`GetDiskFreeSpaceEx`) y `Documentos()` (carpeta conocida `FOLDERID_Documents`, que respeta las redirecciones).

- [ ] **Step 1: Pruebas** de tabla para `Ajuste`:
  - (8192, 4) → (4096, 2, false)
  - (4096, 2) → (2048, 1, true)
  - (16384, 16) → (4096, 4, false)
  - (6000, 2) → (2560, 1, true)

  Y `EspacioNecesario`: para 3,5 GB da 7 GB + 12 GiB.
- [ ] **Steps 2–4:** fallar, implementar, pasar.
- [ ] **Step 5: Commit**: `feat(instalador): ajuste de memoria y espacio`.

### Task B7: Flujo, ventanas y acceso directo

**Files:** `instalador/main.go`, `instalador/internal/flujo/flujo.go`, `flujo_test.go`, `instalador/internal/ui/ui.go` (zenity), `instalador/internal/acceso/acceso_windows.go` (go-ole: `WScript.Shell.CreateShortcut`) e `instalador/internal/registro/registro.go`

**Interfaces:**
- Consumes: todo lo de B2 a B6.
- `type UI interface { Aviso(texto string); Error(texto string); Pregunta(texto, si, no string) bool; Progreso(titulo string) Barra }`.
- `flujo.Ejecutar(ctx, dep flujo.Dependencias) error`. `Dependencias` agrupa `UI`, `vbox.Ejecutor`, `localizar`, el cliente HTTP, `equipo` y las rutas, para poder sustituirlo todo en las pruebas.
- Carpetas:
  - trabajo: `%LOCALAPPDATA%\MintDocente\descarga\`;
  - registro: `%LOCALAPPDATA%\MintDocente\instalador.log`;
  - Intercambio: `<Documentos>\Intercambio`.

**Orden del flujo (del spec):**
1. Localizar VirtualBox (`ErrNoInstalado` → mensaje fijado).
2. Comprobar la versión (menor que `VBOX_MINIMA` → mensaje fijado).
3. Comprobar si la máquina existe.
   - Si existe y está encendida: «Mint Docente está abierta. Ciérrala y vuelve a abrir el instalador.».
   - Si existe y está apagada: preguntar «Abrir» o «Reinstalar». Reinstalar lleva el aviso de que se borrará lo que no esté guardado, y una segunda confirmación antes de `Borrar`.
4. Leer `SHA256SUMS` y comprobar el espacio en la unidad de la carpeta de máquinas. Si falta: «Faltan N GB libres en el disco C:.».
5. Descargar con barra de progreso.
6. Unir.
7. Importar con `Ajuste`. Si `lenta`, avisar una vez de que el equipo tiene poca memoria y la máquina irá lenta.
8. Crear Intercambio y compartirla. Si falla, se avisa de que no habrá carpeta Intercambio (que use arrastrar y soltar) y se sigue.
9. Acceso directo «Mint Docente» en el escritorio, que apunta a `VirtualBoxVM.exe --startvm "Mint Docente"`.
10. Borrar la carpeta de descarga.
11. Arrancar.

**Errores:**
- Cada error se muestra con una frase y lo que se puede hacer, y añade: «Si necesitas ayuda, envía el archivo instalador.log de la carpeta MintDocente.» (con la ruta). El detalle técnico va solo al registro.
- Si la importación falla a medias, se borra la máquina parcial con `Borrar`.
- Una descarga a medias se conserva para reanudar.

- [ ] **Step 1: Pruebas del flujo** con dependencias falsas:
  - `TestSinVirtualBox`: el mensaje exacto y ninguna llamada a `import`.
  - `TestVirtualBoxAntiguo`.
  - `TestMaquinaEncendida`: no se llama a `unregistervm`.
  - `TestReinstalarPideDosVeces`: con un «no» en la segunda pregunta, no se llama a `unregistervm`.
  - `TestSinEspacio`: el mensaje contiene los GB que faltan y no hay descarga.
  - `TestImportFallaLimpia`: se llama a `unregistervm --delete` y se muestra un error.
  - `TestIntercambioFallaSigue`: se llama a `startvm` y se muestra el aviso.
  - `TestCaminoFeliz`: el orden de las órdenes es `import`, `sharedfolder` y `startvm`, y la carpeta de descarga desaparece.
  - `TestDescargaCortadaSeReanuda`: un primer `Ejecutar` con la red cortada deja `.descargando`; el segundo continúa.
- [ ] **Steps 2–4:** fallar, implementar, pasar (`go test ./...`).
- [ ] **Step 5:** `instalador/construir.sh` y prueba manual en el Windows de B1 con una cuenta estándar: el camino feliz, la reinstalación, cortar la red a mitad y repetir, y un usuario con tildes.
- [ ] **Step 6: Commit**: `feat(instalador): flujo completo con ventanas y acceso directo`.

### Task B8: Release, página de descarga y lista para el centro

**Files:**
- Create: `preparar-release.sh`: a partir de `salida/MintDocente-AAAAMMDD.ova` y `salida/InstalarMintDocente.exe`, crea `salida/release-<etiqueta>/` con las partes `.ova.partN` (de 1 992 294 400 bytes, como en v1.1), `SHA256SUMS` (partes + completa) y el `.exe`. **No sube nada**: subirlo con `gh release create` lo hace el usuario.
- Create: `docs/prueba-en-el-centro.md`: lista de comprobación para el usuario, con los puntos del spec («Lo que solo se sabrá en el centro») y qué anotar en cada uno.
- Modify: `README.md`: la descarga pasa a ser «Descarga `InstalarMintDocente.exe` y ábrelo», con la captura de SmartScreen de B1. El método manual se conserva para macOS y Linux. Incluye cómo preparar una release.
- Modify: `files/GUIA-DOCENTE.md`: «Primer arranque» empieza en el instalador.

- [ ] **Step 1: Prueba de `preparar-release.sh`** con una OVA falsa de 5 MB y un tamaño de parte pequeño (variable `TAM_PARTE`): las partes concatenadas son idénticas al original, `SHA256SUMS` pasa `sha256sum -c` y lo lee el `sumas.Leer` de B2 (`go test` con ese archivo como dato).
- [ ] **Step 2: Implementar y pasar.**
- [ ] **Step 3: Redactar** el README, la guía y la lista.
- [ ] **Step 4: Commit**: `feat: preparar releases con el instalador y guía de descarga`.

---

## Al terminar

- Revisión final de toda la rama.
- Entregar al usuario:
  - el tamaño final de la OVA;
  - la carpeta `salida/release-<etiqueta>/` lista para subir;
  - `docs/prueba-en-el-centro.md` para su prueba en el sobremesa;
  - las decisiones de este plan que no estaban en el spec: el kernel retenido sin actualizaciones automáticas, la regla de memoria y CPU, `VBOX_MINIMA` y la lista de aplicaciones quitadas.

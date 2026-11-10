# Mint Docente

Genera una máquina virtual **llave en mano** para VirtualBox (`.ova`) con:
- Linux Mint 22.3 XFCE en español;
- **OpenCode Desktop v2** como entorno de trabajo;
- las herramientas que necesita un agente de IA para preparar, guiado por un docente, temarios, actividades, rúbricas, exámenes y documentos PDF/DOCX/ODT.

## Descargar

**Windows:** descarga [`InstalarMintDocente.exe`](https://github.com/balejosg/mint-docente/releases/latest/download/InstalarMintDocente.exe) y ábrelo. Necesitas VirtualBox 7 instalado (si no lo tienes, pídeselo al coordinador TIC); el instalador no pide permisos de administrador.

El instalador no está firmado, así que Windows puede mostrar una ventana azul, «Windows protegió su PC». Pulsa **Más información** (debajo del texto) y después **Ejecutar de todos modos**.

Después el instalador lo hace todo solo:
1. Descarga Mint Docente (unos 3 GB) y comprueba que ha llegado entera. Si la descarga se corta, al volver a abrirlo sigue donde lo dejó.
2. La instala en VirtualBox con la memoria y los procesadores ajustados a tu equipo.
3. Crea la carpeta **Intercambio** en tus Documentos, para pasar archivos entre Windows y Mint Docente.
4. Deja en el escritorio un acceso directo **Mint Docente** y la arranca.

Si ya la tienes instalada, te pregunta si quieres abrirla o reinstalarla. Si algo falla, el mensaje dice qué hacer; el registro está en `%LOCALAPPDATA%\MintDocente\instalador.log`.

### Método manual (si el instalador no funciona)

En macOS y Linux es el único método. La OVA está en [Releases](https://github.com/balejosg/mint-docente/releases/latest), dividida en partes de menos de 2 GB (límite de GitHub): `MintDocente-AAAAMMDD.ova.part0`, `.part1`, `.part2`… Descarga **todas** las partes y el `SHA256SUMS` en la misma carpeta. El nombre de la `.ova` es el de la última línea de `SHA256SUMS`; ponlo en la primera línea de lo que sigue.

Las partes se unen **en orden numérico** (`part0`, `part1`, …, `part9`, `part10`…). No uses `part*`: después de `part9` ese patrón las desordena (`part10` iría antes que `part2`) y la OVA saldría dañada.

**Windows** (PowerShell, en la carpeta de la descarga):

```powershell
$ova = "MintDocente-AAAAMMDD.ova"
$n = (Get-ChildItem "$ova.part*").Count
$salida = [IO.File]::Create("$PWD/$ova")
for ($i = 0; $i -lt $n; $i++) { $e = [IO.File]::OpenRead("$PWD/$ova.part$i"); $e.CopyTo($salida); $e.Close() }
$salida.Close()
$esperado = ((Get-Content SHA256SUMS | Where-Object { $_ -match "[ *]$([regex]::Escape($ova))$" }) -split '\s+')[0]
if ((Get-FileHash $ova -Algorithm SHA256).Hash -eq $esperado) { "OVA correcta" } else { "La OVA NO coincide: vuelve a descargar las partes" }
```

**macOS y Linux** (Terminal, en la carpeta de la descarga):

```bash
ova=MintDocente-AAAAMMDD.ova
n=$(ls "$ova".part* | wc -l)
for i in $(seq 0 $((n - 1))); do cat "$ova.part$i"; done > "$ova"
grep -E "[ *]$ova\$" SHA256SUMS | shasum -a 256 -c -   # debe decir «OK»
```

Luego impórtala en VirtualBox (*Archivo → Importar servicio virtualizado*).

## Qué encuentra el docente

1. Abre la máquina (en Windows la instala y la arranca el instalador; en macOS y Linux, se importa la `.ova` en VirtualBox). Entra solo, como el usuario `docente` (contraseña `docente`).
2. La primera vez, un asistente le guía para conectar **GitHub Copilot** (inicio de sesión con código en github.com/login/device) u otro proveedor de pago. La imagen no contiene ninguna credencial.
3. Se abre OpenCode sobre su carpeta de trabajo, `Documentos/Proyecto predeterminado`. Es la que crea OpenCode Desktop en su primer arranque; no se puede elegir otra. Está vacía, la organizan el docente y el agente, y en el escritorio aparece como **Aula**.

## Qué incluye

| Área | Herramientas |
|---|---|
| Office | LibreOffice (es), pandoc, fuentes de Microsoft + Carlito/Caladea |
| PDF | poppler-utils, qpdf, ghostscript, img2pdf, PDF Arranger, Xournal++ |
| OCR | ocrmypdf + tesseract (spa, cat, glg, eus, eng) |
| Fórmulas | Typst 0.15, TeX Live (recommended, extra, xetex, español) |
| Python (`/opt/docente-py`) | python-docx, openpyxl, python-pptx, pypdf, pdfplumber, odfpy, reportlab, markitdown |
| Normativa de Madrid | Paquete `normativa-madrid` 0.1.0 (wheel con sha256) instalado en `/opt/docente-py`, con la orden `normativa-madrid` para consultar, buscar y comprobar códigos sin Internet. Cobertura: Primaria completa; solo algunas materias de ESO y Bachillerato (computación y digitalización); FP (DAM y algunos módulos optativos). Una vista `.md` por ficha del catálogo en `/usr/share/mint-docente/normativa/` |
| OpenCode | `/etc/opencode/opencode.jsonc` (permisos, sin autoactualización), `AGENTS.md` y 7 skills en español en `/usr/share/mint-docente/` |
| Sistema | Un solo kernel, retenido (`apt-mark hold`) y sin actualizaciones automáticas: no hay compiladores (ni `build-essential` ni DKMS) para recompilar las Guest Additions. Sin Thunderbird, HexChat, Transmission, Hypnotix, Celluloid, Rhythmbox, Warpinator ni Timeshift. Sin firmware, documentación ni manuales; idiomas: es, ca, gl, eu y en (`/etc/dpkg/dpkg.cfg.d/90-mint-docente` hace que las actualizaciones tampoco los traigan). Java se conserva para LibreOffice |
| Compartir (opcional) | GitHub CLI 2.103.3 (sha256), la orden `compartir` y la skill `compartir`: espacios compartidos entre docentes en repositorios privados de su cuenta de GitHub (invitar, quitar acceso, traer, guardar sin perder versiones, usar skills de compañeros) y «Mi trabajo» (`compartir mi-trabajo guardar|traer`: copia privada de toda la carpeta de trabajo y de las skills propias, para seguir en otro equipo, por ejemplo uno del centro que se borra al cerrar la sesión; el agente lo ofrece al empezar con la carpeta vacía y solo guarda cuando se lo piden). El menú «Desconectar mis cuentas» cierra las sesiones de Copilot y GitHub en un equipo compartido. La sesión de GitHub se inicia solo si el docente quiere, desde el menú «Conectar GitHub para compartir» |

Permisos del agente: tiene libertad dentro de la carpeta de trabajo. Pide confirmación para salir de ella (salvo Intercambio, las plantillas y `/tmp`), para usar `rm` y para las órdenes de `compartir` que crean, invitan, quitan, aceptan, suben (también `mi-trabajo guardar`) o activan una skill. `sudo`, `gh` y `git push` están denegados. Estas reglas son confirmaciones para el docente, no una barrera de seguridad: un agente decidido podría saltárselas por otro camino.

## Construir

Requisitos en el anfitrión de construcción: VirtualBox 7.2, Packer ≥ 1.11, ~60 GB libres, 4 GB de RAM libres y conexión a Internet.

```bash
./build.sh          # → salida/MintDocente-AAAAMMDD.ova (+ .sha256)
VER=1 ./build.sh    # igual, pero mostrando la VM
DEPURAR=1 ./build.sh  # si falla, deja la VM encendida para inspeccionarla
```

Antes de Packer, `build.sh` comprueba en el anfitrión que `files/opencode/skills/normativa-madrid/SKILL.md` tiene el sha256 de `ORIGEN` (es una copia de la skill del paquete). Si no coincide, el build se detiene. `test/probar-normativa.sh` prueba el bloque de instalación de la normativa sin VM (con red): `mal-sha` como argumento debe fallar.

Tarda unos 75 minutos: 12 de instalación y el resto de aprovisionamiento. Al final, `test/verificar.sh` comprueba dentro de la VM:
- las conversiones, el OCR, Typst y las plantillas;
- la carpeta de trabajo;
- tras el recorte y un reinicio: las Guest Additions (vboxsf, portapapeles), un solo kernel retenido, las fuentes, los idiomas y Java;
- que no queda ninguna credencial.

Si alguna prueba falla, el build falla.

Anfitrión Linux con kernel ≥ 6.12: si KVM está cargado, VirtualBox no puede arrancar VMs. `instalar-virtualbox-host.sh` instala VirtualBox 7.2 y añade `kvm.enable_virt_at_load=0`.

### Instalador de Windows

```bash
instalador/construir.sh   # → salida/InstalarMintDocente.exe (descarga Go con versión y sha256 fijados, y pasa las pruebas)
```

Es un programa en Go (`instalador/`) que se compila desde Linux. Lee `https://github.com/balejosg/mint-docente/releases/latest/download/SHA256SUMS` y descarga de ahí las partes. Por eso cada release debe llevar las partes, el `SHA256SUMS` y el `.exe`. La versión mínima de VirtualBox está en `VBOX_MINIMA` (`instalador/internal/flujo/flujo.go`).

### Preparar una release

1. `./build.sh` y `instalador/construir.sh`.
2. `./preparar-release.sh v1.2` (la etiqueta que vayas a publicar). Toma la OVA más reciente de `salida/` (se le puede pasar otra como segundo argumento), la comprueba con su `.sha256` y crea `salida/release-v1.2/` con:
   - las partes `MintDocente-AAAAMMDD.ova.part0`, `.part1`… de 1 992 294 400 bytes (`TAM_PARTE` lo cambia);
   - `SHA256SUMS`: una línea por parte y una para la `.ova` completa. El instalador lee este formato; no lo cambies;
   - `InstalarMintDocente.exe`.

   Si la OVA pasa de 3,8 GB, avisa: revísala antes de publicar. El script no sube nada.
3. Súbela tú: `gh release create v1.2 salida/release-v1.2/* --title "Mint Docente v1.2"`. Tiene que quedar como *latest*, porque el instalador descarga siempre de la última release.
4. Comprueba la descarga con el instalador en un Windows real (ver `docs/prueba-en-el-centro.md`).

## Estructura

```
build.sh                    descarga el ISO (sha256) y ejecuta Packer
packer/mint-docente.pkr.hcl VM (2 CPU, 4 GB, 40 GB, VMSVGA) y exportación OVA
packer/fuentes/preseed.cfg.in   plantilla de la instalación desatendida (Ubiquity)
packer/fuentes/postinstalacion.sh  SSH temporal para Packer (se incrusta en el preseed)
packer/generar-preseed.sh   genera packer/http/preseed.cfg (lo llama build.sh)
provision/00-base.sh        Guest Additions, actualización, idioma, sin bloqueo de pantalla
provision/10-documentos.sh  herramientas de documentos y plantillas
provision/20-opencode.sh    OpenCode (versión fijada con sha256), configuración, skills
provision/30-escritorio.sh  autoarranque, asistente, lanzadores, guía PDF
provision/80-recorte.sh     imagen más ligera: kernel retenido, sin compiladores ni programas sobrantes, reinicio
provision/90-limpieza.sh    autologin, sin SSH ni credenciales, puesta a cero del disco
files/                      configuración, skills, plantillas, scripts del escritorio
files/compartir/compartir   orden para compartir espacios entre docentes (Python, sin dependencias)
test/verificar.sh           pruebas dentro de la VM
test/probar-normativa.sh    prueba en el anfitrión el bloque de normativa-madrid (venv temporal)
test/test_compartir.py      pruebas de compartir sin red (gh falso: test/falso-gh); `python3 test/test_compartir.py`
instalador/                 instalador de Windows en Go (construir.sh lo compila)
preparar-release.sh         parte la OVA y prepara salida/release-<etiqueta>/ (sin subir nada)
docs/prueba-en-el-centro.md lista de pruebas en un equipo real del centro
docs/superpowers/          notas históricas del currículo propio, ya sustituido por normativa-madrid
```

## Actualizar versiones

- **OpenCode:** `OPENCODE_VERSION` y `OPENCODE_DEB_SHA256` en `provision/20-opencode.sh`.
- **GitHub CLI:** `GH_VERSION` y `GH_DEB_SHA256` en `provision/20-opencode.sh` (el sha256 está en `gh_<versión>_checksums.txt` de la release).
- **Mint:** `ISO_URL`/`ISO_SHA256` en `build.sh` y `casper_uuid` en el `.pkr.hcl` (está en `isolinux/live.cfg` del ISO).
- **Typst:** `TYPST_VERSION`/`TYPST_SHA256` en `provision/10-documentos.sh`.
- **Normativa de Madrid:** los datos y el código viven en el repositorio del paquete (`normativa-madrid`), no aquí. Para subir de versión:
  1. Publica la release en ese repositorio.
  2. Cambia el número de versión en la tabla de este README («0.1.0»).
  3. Cambia la URL y `NORMATIVA_WHEEL_SHA256` en `provision/20-opencode.sh` (y el número de versión en su título).
  4. Copia `skill/normativa-madrid/SKILL.md` de la nueva etiqueta a `files/opencode/skills/normativa-madrid/` y actualiza la URL y el sha256 de `ORIGEN`.
  5. Contrasta con `normativa-madrid catalogo` la cobertura descrita en `files/opencode/AGENTS.md`, `files/GUIA-DOCENTE.md`, este README y `files/opencode/skills/temario/SKILL.md`.
  6. Ejecuta `test/probar-normativa.sh`.
- **Cada curso:** el paquete toma el curso académico de la fecha de hoy y la 0.1.0 solo trae datos de 2026-27. A partir del 1-9-2027 las consultas por defecto darían `CURSO_SIN_DATOS`, así que la OVA hay que reconstruirla cada curso con un paquete que traiga los datos del curso nuevo (el agente tiene una regla de reserva con `--curso-academico`, pero con datos antiguos).

## Notas técnicas (aprendidas al construirla)

- **La instalación no tiene red.** casper desactiva la red tras descargar el preseed, y debconf rompe las órdenes con `\` o comillas dobles. Por eso `success_command` no instala nada: descomprime un script embebido (gzip+base64) que crea un servicio. Ese servicio instala openssh-server en el primer arranque del sistema instalado.
- **OpenCode Desktop v2 no puede abrir una carpeta elegida.** En su primer arranque crea y abre `Documentos/Proyecto predeterminado`. Por eso esa es la carpeta de trabajo y el icono «Aula» apunta a ella. Su estado (`~/.config/ai.opencode.desktop`) se borra al final del build para que ese primer arranque ocurra en el equipo del docente.
- **Sin cuenta conectada, OpenCode usa modelos gratuitos de OpenCode Zen** (p. ej. *Big Pickle*), así que funciona aunque el docente omita el asistente.
- **Asistente de Copilot:** `opencode auth login github-copilot --method device --answer deploymentType=github.com` imprime `Enter code: XXXX-XXXX` sin preguntar nada. El asistente lee ese código y lo muestra.
- **`opencode run` sin terminal lee la entrada estándar:** para usarlo en scripts, añade `</dev/null`.
- **Las Guest Additions se compilan una vez, en el build.** `80-recorte.sh` quita los compiladores y retiene el metapaquete del kernel: un kernel nuevo dejaría sin carpeta compartida, portapapeles ni ajuste de pantalla. Para actualizar el kernel, se reconstruye la OVA.

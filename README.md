# Mint Docente

Genera una máquina virtual **llave en mano** para VirtualBox (`.ova`) con:
- Linux Mint 22.3 XFCE en español;
- **OpenCode Desktop v2** como entorno de trabajo;
- las herramientas que necesita un agente de IA para preparar, guiado por un docente, temarios, actividades, rúbricas, exámenes y documentos PDF/DOCX/ODT.

## Qué encuentra el docente

1. Importa la `.ova` en VirtualBox y la arranca. Entra solo, como el usuario `docente` (contraseña `docente`).
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
| OpenCode | `/etc/opencode/opencode.jsonc` (permisos, sin autoactualización), `AGENTS.md` y 5 skills en español en `/usr/share/mint-docente/` |

Permisos del agente: tiene libertad dentro de la carpeta de trabajo. Pide confirmación para salir de ella (salvo Intercambio, las plantillas y `/tmp`) y para usar `rm`. `sudo` está denegado.

## Construir

Requisitos en el anfitrión de construcción: VirtualBox 7.2, Packer ≥ 1.11, ~60 GB libres, 4 GB de RAM libres y conexión a Internet.

```bash
./build.sh          # → salida/MintDocente-AAAAMMDD.ova (+ .sha256)
VER=1 ./build.sh    # igual, pero mostrando la VM
DEPURAR=1 ./build.sh  # si falla, deja la VM encendida para inspeccionarla
```

Tarda unos 75 minutos: 12 de instalación y el resto de aprovisionamiento. Al final, `test/verificar.sh` comprueba dentro de la VM:
- las conversiones, el OCR, Typst y las plantillas;
- la carpeta de trabajo;
- que no queda ninguna credencial.

Si alguna prueba falla, el build falla.

Anfitrión Linux con kernel ≥ 6.12: si KVM está cargado, VirtualBox no puede arrancar VMs. `instalar-virtualbox-host.sh` instala VirtualBox 7.2 y añade `kvm.enable_virt_at_load=0`.

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
provision/90-limpieza.sh    autologin, sin SSH ni credenciales, puesta a cero del disco
files/                      configuración, skills, plantillas, scripts del escritorio
test/verificar.sh           pruebas dentro de la VM
```

## Actualizar versiones

- **OpenCode:** `OPENCODE_VERSION` y `OPENCODE_DEB_SHA256` en `provision/20-opencode.sh`.
- **Mint:** `ISO_URL`/`ISO_SHA256` en `build.sh` y `casper_uuid` en el `.pkr.hcl` (está en `isolinux/live.cfg` del ISO).
- **Typst:** `TYPST_VERSION`/`TYPST_SHA256` en `provision/10-documentos.sh`.

## Notas técnicas (aprendidas al construirla)

- **La instalación no tiene red.** casper desactiva la red tras descargar el preseed, y debconf rompe las órdenes con `\` o comillas dobles. Por eso `success_command` no instala nada: descomprime un script embebido (gzip+base64) que crea un servicio. Ese servicio instala openssh-server en el primer arranque del sistema instalado.
- **OpenCode Desktop v2 no puede abrir una carpeta elegida.** En su primer arranque crea y abre `Documentos/Proyecto predeterminado`. Por eso esa es la carpeta de trabajo y el icono «Aula» apunta a ella. Su estado (`~/.config/ai.opencode.desktop`) se borra al final del build para que ese primer arranque ocurra en el equipo del docente.
- **Sin cuenta conectada, OpenCode usa modelos gratuitos de OpenCode Zen** (p. ej. *Big Pickle*), así que funciona aunque el docente omita el asistente.
- **Asistente de Copilot:** `opencode auth login github-copilot --method device --answer deploymentType=github.com` imprime `Enter code: XXXX-XXXX` sin preguntar nada. El asistente lee ese código y lo muestra.
- **`opencode run` sin terminal lee la entrada estándar:** para usarlo en scripts, añade `</dev/null`.

## Pendiente

- Probar el inicio de sesión real con una cuenta de **GitHub Copilot Pro para educadores**.
- Elegir el modelo por defecto de Copilot: idealmente uno que no consuma peticiones *premium*. Se fijaría con `"model"` en `files/opencode/opencode.jsonc`.

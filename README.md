# Mint Docente

Genera una máquina virtual **llave en mano** para VirtualBox (`.ova`) con:
- Linux Mint 22.3 XFCE en español;
- **OpenCode Desktop v2** como entorno de trabajo;
- las herramientas que necesita un agente de IA para preparar, guiado por un docente, temarios, actividades, rúbricas, exámenes y documentos PDF/DOCX/ODT.

## Qué encuentra el docente

1. Importa la `.ova` en VirtualBox y la arranca. Entra solo, como el usuario `docente` (contraseña `docente`).
2. La primera vez, un asistente le guía para conectar **GitHub Copilot** (inicio de sesión con código en github.com/login/device) u otro proveedor de pago. La imagen no contiene ninguna credencial.
3. Se abre OpenCode sobre `~/Aula`, una carpeta vacía que organizan el docente y el agente.

## Qué incluye

| Área | Herramientas |
|---|---|
| Office | LibreOffice (es), pandoc, fuentes de Microsoft + Carlito/Caladea |
| PDF | poppler-utils, qpdf, ghostscript, img2pdf, PDF Arranger, Xournal++ |
| OCR | ocrmypdf + tesseract (spa, cat, glg, eus, eng) |
| Fórmulas | Typst 0.15, TeX Live (recommended, extra, xetex, español) |
| Python (`/opt/docente-py`) | python-docx, openpyxl, python-pptx, pypdf, pdfplumber, odfpy, reportlab, markitdown |
| OpenCode | `/etc/opencode/opencode.jsonc` (permisos, sin autoactualización), `AGENTS.md` y 5 skills en español en `/usr/share/mint-docente/` |

Permisos del agente: tiene libertad dentro de `~/Aula`. Pide confirmación para salir de ella (salvo Intercambio, las plantillas y `/tmp`) y para usar `rm`. `sudo` está denegado.

## Construir

Requisitos en el anfitrión de construcción: VirtualBox 7.2, Packer ≥ 1.11, ~60 GB libres, 4 GB de RAM libres y conexión a Internet.

```bash
./build.sh          # → salida/MintDocente-AAAAMMDD.ova (+ .sha256)
VER=1 ./build.sh    # igual, pero mostrando la VM
```

Al final del build, `test/verificar.sh` comprueba las conversiones, el OCR, Typst, las plantillas, que no queda ninguna credencial, etc. Si alguna prueba falla, el build falla.

## Estructura

```
build.sh                    descarga el ISO (sha256) y ejecuta Packer
packer/mint-docente.pkr.hcl VM (2 CPU, 4 GB, 40 GB, VMSVGA) y exportación OVA
packer/http/preseed.cfg     instalación desatendida de Ubiquity
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

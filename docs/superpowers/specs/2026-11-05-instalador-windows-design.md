# Instalador de un clic para Windows, OVA más ligera y «Mi trabajo» en GitHub

Fecha: 2026-11-05 · Estado: pendiente de aprobación

## Por qué

Hoy, para tener Mint Docente, un docente tiene que descargar tres partes de una OVA de 4,85 GB, unirlas con `copy /b` en la consola de Windows, comprobar un SHA256 e importarla en VirtualBox. Para pasar archivos, además, tiene que crear a mano una carpeta compartida. Muchos docentes no técnicos se quedan en este paso, antes de llegar a la parte cuidada de la máquina (el asistente de primer arranque, el agente sin jerga).

Queremos que **instalar Mint Docente sea un doble clic**, también en los equipos del centro.

## Para quién y en qué equipos (decisiones del usuario)

- Docentes con **Windows**: en su **portátil personal** o en un **sobremesa del centro**.
- **VirtualBox ya está instalado.** El docente **no tiene permisos de instalación** (no puede elevar privilegios).
- En los sobremesas del centro **puede que no se conserve nada** entre sesiones: el perfil se borra o el equipo se restaura. Llevar un disco USB o usar una unidad de red del centro **no son opciones realistas**.
- El trabajo se salva **en GitHub**, con la orden `compartir` que ya existe. Drive queda descartado.
- El asistente **guarda solo cuando el docente se lo pide**, nunca por su cuenta.
- La imagen se **recorta hasta unos 3,5 GB** sin cambiar de distribución. Se descartó cambiar Mint por una base mínima para bajar de 2 GB.

## Objetivos

1. **Instalador de un clic para Windows**, que funciona sin permisos de administrador.
2. **Una OVA más ligera** (~3,5 GB), porque en el centro se descarga en cada sesión.
3. **«Guarda mi trabajo» / «Trae mi trabajo»**: el docente se lo pide al asistente y su trabajo va y vuelve de GitHub, también a una máquina recién instalada.

## Lo que no se hace

- macOS ni Linux como anfitriones (siguen con el método manual del README). Los Mac con Apple Silicon no pueden ejecutar esta OVA x86 en ningún caso.
- Instalar VirtualBox.
- Integración con Drive, OneDrive o la nube de EducaMadrid.
- Guardado automático.
- Cambiar la distribución base.
- El perfil del docente y los encargos de ejemplo: van en un proyecto aparte.

## Lo que lo haría un fracaso aunque funcione

- Que el docente tenga que abrir una consola, escribir una orden o leer un mensaje técnico.
- Que en un sobremesa del centro el instalador pida permisos de administrador o lo bloqueen las directivas del equipo sin explicar qué pasa.
- Que se pierda trabajo: al reinstalar sobre una máquina existente, o al traer el trabajo encima de algo más reciente.
- Que se suban claves a GitHub, o datos del alumnado sin el visto bueno explícito del docente.
- Que el recorte de la imagen rompa algo que hoy funciona (conversiones, OCR, fórmulas, las Guest Additions, la carpeta compartida).

## Parte A — OVA más ligera

Medición de la OVA v1.1 (2026-11-05): hay 10,3 GB de archivos en el disco y se comprimen a 4,85 GB. El `swapfile` (4 GB) está casi vacío y no influye en la descarga.

Candidatos a quitar (MB sin comprimir):

| Qué | MB | Nota |
|---|---|---|
| `/usr/lib/firmware` | 684 | Firmware de hardware real; en una VM no se usa |
| Temas de iconos | ~600 de 914 | Conservar el tema en uso y sus dependencias |
| Fuentes Noto de otros alfabetos | ~400 de 495 | Conservar latín, griego, cirílico y símbolos o matemáticas |
| Thunderbird | 333 | |
| Kernel 6.14 (módulos, initrd y vmlinuz) | 265 | Quedó del ISO; se usa el 7.0 |
| `/usr/src` (cabeceras), `build-essential`, `dkms` | 216+ | Ver la restricción de las Guest Additions |
| Documentación, `locale` y `man` | ~300 | Conservar es, ca, gl, eu y en |
| Fondos de escritorio | 81 | Conservar el que se usa |

**Meta:** una OVA de unos 3,5 GB. Seguirá en 2 partes en GitHub Releases, pero el instalador las une.

**Restricciones:**
- `test/verificar.sh` tiene que seguir pasando entero. Si el recorte afecta a algo que no cubre (por ejemplo la carpeta compartida o el portapapeles), se añade una prueba.
- **Guest Additions:** están compiladas con DKMS. Si se quitan las cabeceras y el compilador y luego se actualiza el kernel, dejan de funcionar la carpeta compartida, el portapapeles y el ajuste de pantalla. Antes de quitarlos hay que comprobar si la máquina puede actualizar el kernel (gestor de actualizaciones de Mint, actualizaciones desatendidas). Si puede, o se bloquean las actualizaciones del kernel o se conservan las cabeceras.
- LibreOffice no debe perder funciones que use el docente al quitar o conservar Java (`/usr/lib/jvm`, 203 MB). Lo decide quien lo construya, con una prueba que lo respalde.

## Parte B — Instalador para Windows

**Cómo llega:** en la página de descarga (README y Releases) hay un archivo pequeño, «Instalar Mint Docente». Se descarga y se le hace doble clic.

**Qué hace, en orden:**
1. Comprueba que VirtualBox está instalado. Si no, lo dice en una frase («Este equipo no tiene VirtualBox. Pídeselo al coordinador TIC.»).
2. Comprueba el espacio libre: descarga, OVA unida y disco importado. Si no cabe, dice cuánto falta.
3. Si ya existe una máquina «Mint Docente» del usuario, **no la pisa**: pregunta si abrirla o reinstalarla y avisa de que reinstalar borra lo que no esté guardado.
4. Descarga las partes de la última release, con progreso visible y reanudación si se corta.
5. Une las partes y comprueba el SHA256 con `SHA256SUMS`. Si falla, vuelve a descargar la parte afectada o lo explica.
6. Importa la OVA en la carpeta de VirtualBox del usuario, sin administrador.
7. **Ajusta la memoria y los procesadores** a lo que tenga el equipo. Hoy la OVA pide 4 GB y 2 CPU, y un sobremesa con 8 GB puede ir justo. Quien lo construya decide la regla.
8. Crea la carpeta **Intercambio** en los Documentos de Windows del usuario y la monta automáticamente en la máquina.
9. Deja un acceso directo «Mint Docente» en el escritorio de Windows y arranca la máquina.
10. Borra la descarga y la OVA unida, que ya no hacen falta.

**Mensajes:** todo en español llano y en ventanas, nunca en una consola que el docente tenga que leer. Los errores dicen qué ha pasado y qué hacer.

**Formato (decisión del usuario):** un **`.exe` sin firmar**. Se acepta el aviso de SmartScreen: la página de descarga explica, con una captura, «Más información → Ejecutar de todos modos». Los scripts `.ps1` o `.bat` quedan descartados como formato principal, porque la directiva de ejecución o AppLocker los bloquean con más facilidad.

**Cómo se valida (decisión del usuario):** nadie puede probar antes en un sobremesa del centro. **El usuario probará la versión definitiva** en uno. Por eso, quien lo construya:
- hace antes un spike propio en un Windows con una **cuenta de usuario estándar** (sirve una máquina virtual con Windows y VirtualBox), para confirmar que el `.exe` y `VBoxManage import` funcionan sin administrador;
- diseña el instalador **a la defensiva** para lo que solo se sabrá en el centro: cada paso detecta su fallo y lo explica en llano (directivas que bloquean, VirtualBox antiguo, poco espacio, ruta de red), en lugar de suponer que funcionará;
- **comprueba la versión de VirtualBox (decisión mía).** Si es anterior a la que admite la OVA, el instalador lo dice («Este VirtualBox es antiguo; pide al coordinador TIC la versión 7.x») en lugar de fallar al importar. La versión mínima la fija quien lo construya, probando la importación;
- entrega al usuario una **lista corta de comprobación** para la prueba en el centro, con los puntos siguientes.

Lo que solo se sabrá en el centro:
- si el `.exe` sin firmar se puede ejecutar (SmartScreen, AppLocker);
- si `VBoxManage import` funciona como usuario sin administrador;
- dónde está `VBoxManage.exe` (en el registro o en `Program Files`);
- si la carpeta compartida automontada funciona sin permisos;
- cuánto tarda la descarga y la importación con la red del centro;
- la versión de Windows y de VirtualBox de los centros. La OVA se exporta hoy con VirtualBox 7.2 (VMSVGA, xHCI); si los centros tienen una versión anterior, puede que haya que cambiar la configuración en Packer;
- si los perfiles son móviles o tienen Documentos redirigido a una unidad de red con cuota, y si hay algún disco local con permiso de escritura fuera del perfil. De eso depende dónde van la descarga, el disco de la máquina (~10 GB) e Intercambio;
- si los docentes comparten una cuenta genérica de Windows. En ese caso, el siguiente usuario vería la máquina con las sesiones de Copilot y GitHub del anterior (ver la Parte C).

Si en el centro las directivas bloquean el `.exe`, hay que volver a hablarlo con el usuario: este diseño no tiene alternativa para ese caso.

## Parte C — «Guarda mi trabajo» / «Trae mi trabajo»

Se apoya en `compartir` y en la skill `compartir` (ver `files/compartir/compartir` y `files/opencode/skills/compartir/SKILL.md`). Hoy `compartir` solo gestiona espacios dentro de `Aula/Compartido`. En una máquina nueva no sabe recuperar de golpe los espacios propios ni las skills de compañeros en uso.

**Comportamiento esperado:**
- **«Guarda mi trabajo»:** el asistente guarda la carpeta Aula entera (salvo `Compartido/`) en un **espacio privado del docente, «Mi trabajo»**, sin compañeros. También guarda los espacios compartidos con cambios. Antes de subir aplica las mismas revisiones que `compartir guardar`: datos del alumnado, material con derechos y claves. **Datos del alumnado (decisión del usuario):** igual que hoy, se suben solo con el visto bueno explícito del docente, después de avisarle. Así no se pierden en el centro. Las claves nunca se suben.
- **«Trae mi trabajo»** en una máquina recién instalada: recupera «Mi trabajo» en el Aula, vuelve a descargar los espacios propios y los aceptados, y vuelve a activar las skills de compañeros que el docente usaba. Para eso, «Mi trabajo» guarda también la lista de esas skills.
- **Al empezar con el Aula vacía y GitHub conectado**, el asistente ofrece traer el trabajo. Si GitHub no está conectado, ofrece conectarlo primero (el mismo flujo de `mint-docente-github`).
- **No se pierde nada al traer:** si el Aula local tiene algo que no está en «Mi trabajo», se aplica la regla de `dos_versiones` que ya usa `compartir`.
- **Recordatorio (decisión mía, el usuario puede quitarlo):** al terminar cada encargo, el asistente recuerda en una línea que pida «guarda mi trabajo» antes de irse. No guarda por su cuenta.
- Los permisos se mantienen igual: guardar pide confirmación, igual que hoy `compartir guardar`.
- **Qué entra en «Mi trabajo» (decisión mía):** el Aula (salvo `Compartido/`) y las **skills propias del docente** de `~/.config/opencode/skills` (las de compañeros se restauran desde su espacio). **No entra el historial de conversaciones de OpenCode**: puede contener datos del alumnado y no hace falta para seguir trabajando.
- **Archivos que GitHub no admite (decisión mía):** los de más de 95 MB no se suben. El asistente los nombra antes de guardar y avisa de que en un equipo del centro se perderán al cerrar la sesión, salvo que el docente los copie a Intercambio o a otro sitio.
- **Desconectar al irse (decisión mía):** cuando el docente pide guardar porque se va, el asistente le ofrece desconectar sus cuentas de Copilot y GitHub de esa máquina, para el caso de que el equipo no se borre y lo use otra persona.

Lo que no se decide aquí y queda para quien lo construya: si es una orden nueva de `compartir` o un uso de las existentes, el nombre del repositorio de «Mi trabajo» y el formato de la lista de skills.

## Lo que se mantiene

- Ninguna credencial dentro de la imagen.
- En cada máquina nueva, el docente vuelve a conectar Copilot y GitHub con los flujos de código que ya existen.
- La guía (`files/GUIA-DOCENTE.md`) y el README se actualizan: la instalación con el instalador, y guardar y traer el trabajo en el centro.

## Lo que queda a criterio de quien lo construya

- La tecnología con la que se genera el `.exe`, según su spike en Windows con usuario estándar.
- La regla de memoria y CPU.
- El detalle del recorte y las pruebas nuevas.
- El diseño interno de la Parte C.

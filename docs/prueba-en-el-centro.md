# Prueba en el centro

Lista para probar el instalador y la máquina en un equipo real del centro, antes de recomendarlo a los docentes. Hay cosas que solo se saben allí. Cada apartado dice qué hacer, en qué fijarse y qué anotar. Si algo falla, copia el mensaje tal cual y guarda `%LOCALAPPDATA%\MintDocente\instalador.log`.

## Antes de empezar

- Un equipo del centro con Windows 10 u 11 y VirtualBox 7.x, instalado por el coordinador TIC.
- Una **cuenta estándar** (sin administrador). Mejor si el nombre de usuario lleva tildes y un espacio, por ejemplo `José Ángel`, para que la carpeta sea `C:\Users\José Ángel`.
- Una cuenta de GitHub de prueba con Copilot (y, para el apartado 6, un archivo de más de 95 MB, por ejemplo un vídeo).
- `InstalarMintDocente.exe` de la última release.

**Anota primero:**
- versión de Windows (*Configuración → Sistema → Información*);
- versión de VirtualBox (*Ayuda → Acerca de VirtualBox*);
- dónde está `VBoxManage.exe` (normalmente `C:\Program Files\Oracle\VirtualBox\`);
- si los docentes usan una **cuenta genérica** compartida o cada uno la suya;
- si el perfil es **móvil** o se **borra al cerrar la sesión**;
- si **Documentos** está redirigido a una unidad de red (`\\servidor\…`): en el Explorador, clic derecho en Documentos → *Propiedades → Ubicación*. Anota la ruta y si tiene cuota;
- si hay algún disco local con permiso de escritura fuera del perfil (por ejemplo `D:\`), y cuánto espacio libre hay en `C:` y en ese disco.

## 1. Abrir el instalador

1. Haz doble clic en `InstalarMintDocente.exe`.
2. Si sale «Windows protegió su PC», pulsa **Más información** → **Ejecutar de todos modos**.
3. **Haz una captura** de la ventana de SmartScreen antes de pulsar, y otra después de pulsar *Más información*. Son para la guía del docente.

Fíjate en que no aparece ninguna petición de administrador (control de cuentas de usuario) ni ninguna ventana negra de consola.

Anota:
- si SmartScreen lo deja pasar con esos dos clics;
- si **AppLocker** u otra directiva lo bloquea del todo. En ese caso, anota el mensaje exacto y para aquí: hay que hablarlo, porque el diseño no tiene alternativa para ese caso.

## 2. Instalación normal

1. Deja que termine sin tocar nada.
2. Durante la descarga, mira que la barra avanza en GB y que `%LOCALAPPDATA%\MintDocente\descarga` se va llenando.

Al terminar, comprueba:
- la carpeta `descarga` ha desaparecido;
- la máquina está en `%USERPROFILE%\VirtualBox VMs\Mint Docente` (con tildes en la ruta, si el usuario las tiene);
- en VirtualBox, *Configuración → Sistema*, la memoria y los procesadores se han ajustado al equipo;
- existe `Documentos\Intercambio` en Windows;
- en el escritorio hay un acceso directo **Mint Docente** (también si el escritorio está en OneDrive), y abrirlo arranca la máquina;
- la máquina arranca sola al final;
- `instalador.log` se lee bien y las tildes salen bien.

Anota:
- cuánto ha tardado la **descarga** y cuánto la **importación**, con la red del centro;
- si la importación ha funcionado con la cuenta estándar, sin administrador;
- si el centro usa proxy (y si es PAC). Con proxy, la descarga tiene que funcionar igual.

## 3. Dentro de la máquina

Con la máquina recién instalada:
1. **Intercambio:** deja un archivo en `Documentos\Intercambio` de Windows. Debe aparecer en la carpeta **Intercambio** de Mint Docente. Crea otro desde Mint Docente y comprueba que se ve en Windows.
2. **Portapapeles:** copia un texto en Windows y pégalo en Mint Docente, y al revés.
3. **Pantalla:** cambia el tamaño de la ventana de VirtualBox. El escritorio de Mint Docente debe ajustarse solo.
4. **OpenCode** se abre al arrancar (después del asistente de la cuenta de IA).

Anota si la carpeta compartida automontada funciona con la cuenta sin permisos. Si Documentos está en una unidad de red, anota si apareció el aviso «No se ha podido preparar la carpeta Intercambio…» y si la máquina arrancó igual.

## 4. Volver a abrir el instalador

1. **Con la máquina encendida:** debe salir «Mint Docente está abierta. Ciérrala…». No se borra nada.
2. **Con la máquina apagada:**
   - **Abrir** la arranca;
   - **Reinstalar** → **No** no hace nada;
   - cerrar con la X la primera ventana lleva a la segunda pregunta, nunca a borrar;
   - **Reinstalar** → **Sí, reinstalar** la borra y la vuelve a instalar.

## 5. Cuando algo va mal

1. **Corte de red:** durante la descarga, desactiva la red. En unos 30 segundos debe salir «La descarga se ha cortado…», y en la carpeta `descarga` debe quedar un archivo `*.descargando`. Vuelve a activar la red y abre el instalador: la descarga sigue donde se quedó, no desde cero. Compruébalo en la barra y en el registro.
2. **Detener:** pulsa **Detener** durante la descarga. Sale «Has detenido la instalación…». Vuelve a abrirlo: sigue donde lo dejó.
3. **Poco disco:** llena el disco con un archivo grande antes de empezar. Debe salir «Faltan N GB libres en el disco C:.» antes de descargar. Prueba también a llenarlo durante la descarga: sale el mensaje de espacio y lo descargado se conserva.
4. **Sin VirtualBox** (en un equipo que no lo tenga): «Este equipo no tiene VirtualBox. Pídeselo al coordinador TIC.»
5. **VirtualBox antiguo** (6.1, si hay alguno): «Este VirtualBox es antiguo; pide al coordinador TIC la versión 7.x.»
6. **VirtualBox 7.0 o 7.1**, si hay algún equipo con esa versión: comprueba que la máquina se importa y arranca (la OVA se exporta con 7.2). Anota la versión y el resultado. Si no se importa, hay que subir la versión mínima del instalador o cambiar la configuración de la máquina.

## 6. Guardar y traer mi trabajo

Con Copilot y GitHub conectados (cuenta de prueba):
1. Crea un par de archivos en **Aula** y pide «guarda mi trabajo». El asistente revisa los archivos, pide confirmación y guarda.
2. Añade un archivo de más de 95 MB y vuelve a guardar. El asistente lo nombra y avisa de que lo copies a Intercambio.
3. Pídele un encargo cualquiera. Al terminar, recuerda en una línea que pidas «guarda mi trabajo», pero no guarda por su cuenta.
4. Di «guarda mi trabajo, que me voy». El asistente guarda y después ofrece la ventana **Desconectar mis cuentas**. Acepta y comprueba que las sesiones de Copilot y de GitHub quedan cerradas.
5. Comprueba qué borra **Desconectar mis cuentas**: el archivo de credenciales de OpenCode debe ser `~/.local/share/opencode/auth.json` y, tras desconectar, ya no debe estar. Si OpenCode guarda las credenciales en otro sitio, anota cuál.
6. Instala la máquina otra vez (o en otro equipo) y conecta solo Copilot. En el primer mensaje, el asistente pregunta si quieres conectar GitHub. Conéctalo.
7. El asistente ofrece traer tu trabajo. Acepta y comprueba que vuelven el Aula, los espacios compartidos, tus skills y las de compañeros, y que te dice que vuelvas a abrir OpenCode.
8. Con una cuenta de GitHub que nunca haya guardado: el Aula vacía no provoca la oferta, y «trae mi trabajo» responde «Todavía no has guardado…».

Si los docentes comparten una cuenta genérica de Windows, anota si el siguiente usuario vería la máquina con las sesiones del anterior cuando no se desconecta.

## Qué devolver

- Las notas de cada apartado, con los mensajes copiados tal cual.
- Las capturas de SmartScreen.
- `instalador.log` de cada prueba que falle.

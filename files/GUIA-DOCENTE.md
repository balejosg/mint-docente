---
title: "Mint Docente — Guía rápida"
lang: es
geometry: margin=2cm
---

# Qué es

Es un ordenador virtual preparado para trabajar con **OpenCode**, un asistente de inteligencia artificial que prepara contigo **temarios, programaciones, actividades, rúbricas, exámenes** y documentos en **Word (.docx), LibreOffice (.odt) y PDF**.

- **Usuario:** `docente` · **Contraseña:** `docente` (solo se pide para tareas de administración).

# Primer arranque

1. Se abre un asistente que te pide conectar tu cuenta de IA. Elige **GitHub Copilot**.
2. Aparece una ventana con un **código**. Se abre el navegador en `github.com/login/device`: inicia sesión con tu cuenta de GitHub, escribe el código y pulsa **Authorize**.
3. Se abre OpenCode. Ya puedes empezar.

Para cambiar de cuenta más adelante: menú → **Cambiar cuenta de IA**.

# Cómo trabajar

Escribe lo que necesitas como se lo pedirías a un compañero. Por ejemplo:

- «Crea el temario de Biología y Geología de 3º ESO: 8 unidades, en Word y PDF.»
- «Hazme una ficha de 10 actividades sobre la célula, con una versión adaptada.»
- «Prepara un examen tipo test de 20 preguntas del tema 4, con solucionario.»
- «Lee el PDF escaneado que he dejado en Intercambio y conviértelo en un documento editable.»
- «Haz una rúbrica para evaluar una exposición oral.»

El asistente guarda los documentos en tu carpeta **Aula** (está en el escritorio). Si quiere borrar algo o salir de esa carpeta, **te pedirá permiso**.

# Currículo de Madrid

Si das clase en la Comunidad de Madrid, el asistente consulta la normativa oficial (competencias, criterios de evaluación, contenidos, perfil de salida y requisitos de la programación) y la cita tal cual en tus programaciones. Antes de entregarte un documento comprueba que todo lo que cita existe. Si usas criterios de tu libro de texto o de otra comunidad, te propondrá el equivalente de Madrid.

Qué tiene cargado:
- **Primaria:** todas las áreas, con sus competencias, criterios y contenidos por ciclo. Los contenidos de Primaria no tienen numeración oficial en el decreto: el asistente los cita por su bloque y su texto, sin darles un código como si fuera oficial.
- **ESO y Bachillerato:** solo algunas materias, las de computación y digitalización (ESO: Ciencias de la Computación, Tecnología y Digitalización y Digitalización; Bachillerato: Ciencias de la Computación), además del perfil de salida y los requisitos de la programación.
- **Formación Profesional:** el ciclo de Desarrollo de Aplicaciones Multiplataforma y algunos módulos de especialización.

Si das una materia que no está, el asistente te lo dirá: te preparará la programación, pero redactará las competencias y los criterios de forma genérica, sin códigos oficiales. Revísalos con el decreto antes de usarlos.

# Pasar archivos entre tu ordenador y esta máquina

- **Arrastrar y soltar** archivos sobre la ventana de la máquina virtual, o **copiar y pegar** texto.
- **Carpeta compartida** (recomendado, solo se configura una vez). Con la máquina apagada, en VirtualBox:
  1. *Configuración → Carpetas compartidas → +*
  2. Elige una carpeta de tu ordenador, ponle de nombre **Intercambio** y marca **Automontar**.
  3. Arranca la máquina: la carpeta aparece como **Intercambio** en tu carpeta personal.

# Otros programas útiles

- **LibreOffice** para abrir y retocar los documentos.
- **Xournal++** para anotar o firmar PDF.
- **PDF Arranger** para reordenar o unir páginas.

# Consejos

- Revisa siempre lo que genera la IA antes de usarlo con el alumnado.
- No compartas datos personales del alumnado si no es imprescindible.
- Antes de cambios importantes, haz una **instantánea** de la máquina en VirtualBox (*Máquina → Tomar instantánea*).

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
- **Formación Profesional:** el ciclo de Desarrollo de Aplicaciones Multiplataforma y algunos módulos optativos.

Si das una materia que no está, el asistente te lo dirá: te preparará la programación, pero redactará las competencias y los criterios de forma genérica, sin códigos oficiales. Revísalos con el decreto antes de usarlos.

# Pasar archivos entre tu ordenador y esta máquina

- **Arrastrar y soltar** archivos sobre la ventana de la máquina virtual, o **copiar y pegar** texto.
- **Carpeta compartida** (recomendado, solo se configura una vez). Con la máquina apagada, en VirtualBox:
  1. *Configuración → Carpetas compartidas → +*
  2. Elige una carpeta de tu ordenador, ponle de nombre **Intercambio** y marca **Automontar**.
  3. Arranca la máquina: la carpeta aparece como **Intercambio** en tu carpeta personal.

# Compartir con compañeros (opcional)

Puedes compartir con otros docentes tus **aplicaciones, skills, instrucciones (AGENTS.md) y documentos**. Cada cosa que compartes es un **espacio**: una carpeta dentro de *Aula → Compartido* que es igual en tu máquina y en la de los compañeros que invites.

**Conectar una vez.** Menú → **Conectar GitHub para compartir**. Es el mismo código que con Copilot: lo pegas en `github.com/login/device` y pulsas **Authorize**. Si no lo has hecho, el asistente te lo abrirá la primera vez que le pidas compartir algo.

**Pídeselo al asistente**, por ejemplo:

- «Comparte mi skill de rúbricas con Ana; su usuario de GitHub es ana-garcia.»
- «¿Tengo invitaciones? Acepta la de Luis.»
- «Quiero usar la skill de evaluación de Luis.»
- «¿Hay novedades en las skills que uso?»
- «Trae lo último del espacio de 2º de FP» · «Guarda mis cambios en lo compartido».
- «¿Quién tiene acceso a Rúbricas?» · «Quita a Luis.»

**Tu usuario de GitHub:** pregúntaselo al asistente («¿cuál es mi usuario de GitHub?») y dáselo a tus compañeros para que te inviten.

**Ten en cuenta:**

- Quien invitas **puede cambiar y borrar** archivos del espacio. No se pierde nada: GitHub guarda las versiones anteriores.
- Si dos personas cambian lo mismo a la vez, el asistente junta los cambios o te enseña las dos versiones para que elijas.
- **Skills de compañeros:** el asistente te explica qué hacen antes de usarlas, y no cambian solas cuando el compañero las modifica.
- **Claves** (de una IA, de un servicio): no se comparten. Si una aplicación lleva una, el asistente no la sube y te propone cómo apartarla; cada uno pone la suya.
- **No compartas datos del alumnado.** El asistente revisa los archivos antes de subirlos y te avisará, pero la responsabilidad es tuya.

# Otros programas útiles

- **LibreOffice** para abrir y retocar los documentos.
- **Xournal++** para anotar o firmar PDF.
- **PDF Arranger** para reordenar o unir páginas.

# Consejos

- Revisa siempre lo que genera la IA antes de usarlo con el alumnado.
- No compartas datos personales del alumnado si no es imprescindible.
- Antes de cambios importantes, haz una **instantánea** de la máquina en VirtualBox (*Máquina → Tomar instantánea*).

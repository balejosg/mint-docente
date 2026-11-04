# Spike: un solo currículo para Primaria, ESO, Bachillerato y FP (2026-11-04)

Pregunta: ¿se puede sacar `tiza.normativa` a un paquete propio, meter en él la Primaria de Mint Docente y ampliar ESO con el extractor que ya tiene tiza?

Repositorio de tiza en `210d608` (v0.39.0). No he tocado ninguno de los dos repositorios. Pruebas en el scratchpad, con el PDF del Decreto 65/2022 (`bocm-20220726-2`, sha256 correcto, 321 págs.).

## 1. Separar `tiza.normativa`: sí, a bajo coste

- `src/tiza/normativa/` (modelo, datos, consulta y salida; ~1500 líneas) **solo usa la biblioteca estándar** y no importa nada del resto de tiza.
- Fuera del paquete solo lo usan:
  - `cli.py`: `tiza normativa`, `_consulta_normativa`;
  - `ayuda.py` y `aislamiento.py`;
  - siete guías de la skill (`programacion.md`, `unidad.md`, `realinear.md`, `calificador.md`, `dua.md` y `licencias.md`), además de `SKILL.md`.
- Las herramientas del mantenedor (`herramientas/normativa/`, ~3900 líneas) solo dependen de `pdftotext` y de `fuentes.toml` y `lote.toml`.
- Encaja con Mint Docente, cuyo CLI `curriculo` también se escribió solo con la biblioteca estándar.

## 2. El modelo de tiza es mejor que el de Primaria, pero no admite Primaria tal cual

**Lo que tiza tiene y Primaria no:**
- procedencia de cada elemento: norma, lugar y página, además de `via` cuando una norma modifica el texto;
- normas con sha256, correcciones y derogaciones;
- cursos académicos;
- relaciones entre fichas, validadas al cargar.

**Lo que falta en el modelo para Primaria:**

| Hace falta | Situación en tiza |
|---|---|
| Etapa Primaria | `ETAPAS = ("eso", "bachillerato", "fp_gs")` |
| Ficha por **ciclo** (Primaria se organiza en ciclos 1–3; VCE es solo de 5.º; SLE y TYR, de toda la etapa) | `curso` debe cumplir `[1-9]`, o ser `None` |
| Contenidos con **sub-bloque** y viñetas anidadas | `bloque_contenidos` y `contenido` con `padre` (`CO.A.1.1`). Basta con eso: el sub-bloque sería un nivel más |
| Descriptores **propios de Madrid** (Anexo I del Decreto 61/2022) | La ESO usa los estatales (`EST.ESO.DESC`). Para Primaria haría falta una ficha `MAD.EP.DESC` |
| Equivalencia CN/CS ↔ Conocimiento del Medio del RD | No existe. Puede ir en `relaciones` |
| `variantes` de una competencia según el ciclo | Desaparece: con una ficha por ciclo, cada ficha tiene su redacción |
| `curriculo comprobar` (repasa los códigos que cita un documento) | No existe. `tiza normativa elemento <código>` consulta un solo código |

**Diferencia de criterio que hay que decidir:**
- tiza **numera los contenidos** (`MAD.ESO.TECDIG.2.CO.A.1`) aunque el boletín no los numere, y no avisa de que esos códigos no son oficiales.
- En Primaria decidimos lo contrario: los contenidos se citan por bloque y texto literal, sin código.
- Las dos opciones tienen sentido. Un código estable sirve para relacionar elementos y para el calificador, pero no puede presentarse como oficial.

**Códigos:**
- Primaria usa `MAT.1.1.C2`; tiza usa `MAD.ESO.MAT.1.CR.1.1`.
- La OVA de Primaria todavía no ha llegado a los docentes, así que ahora es barato cambiar Primaria a `MAD.EP.MAT.C2.CR.1.1`.

## 3. El extractor de ESO (`extraer_lomloe`) con materias nuevas

Lo he ejecutado sin modificarlo sobre seis materias que tiza no tiene:

| Materia | Resultado |
|---|---|
| Lengua Castellana y Literatura, 1.º | Bien: 10 competencias, 26 criterios y 70 contenidos; descriptores en todas las competencias |
| Geografía e Historia, 1.º | Bien: 9 competencias, 27 criterios y 91 contenidos |
| Educación Física, 1.º | Bien: 5 competencias, 17 criterios y 69 contenidos |
| Matemáticas, 4.º | Bien: 10 competencias, 23 criterios y 75 contenidos |
| Matemáticas, 1.º | **Contenidos mal, y sin dar error.** Solo salen 6 contenidos y el bloque A tiene 1968 caracteres. Este curso usa sub-bloques numerados («1. Conteo.») y viñetas «‒» (U+2012), que el extractor no reconoce, y las mete en el texto del bloque. Los 11 criterios están bien: en 1.º las competencias 4 y 6 no tienen criterios en el boletín. |
| Biología y Geología, 1.º | **Falla y da error:** «no encuentro "Competencias específicas."». El boletín pone «Competencias Específicas.», con mayúscula. |

En ningún texto he encontrado restos de cabecera ni palabras cortadas.

**Conclusión:**
- El extractor sirve para la ESO.
- Cada materia nueva puede traer una variante de maquetación. Algunas dan error, que es lo bueno; otras se cuelan sin avisar, que es lo peligroso.
- Hacen falta los **recuentos esperados por ficha** (`recuentos` en `fuentes.toml`) y una **revisión externa por materia**, como la que se hizo en Primaria.
- Además, el extractor debería rechazar cualquier bloque con viñetas o numeración dentro de su texto.

**Primaria necesita su extractor propio.** `pdftotext -layout` mezcla las columnas del Decreto 61/2022, como ya vimos en el spike de Primaria, así que no sirve. En el paquete común convivirían dos extractores:
- el de tablas con `pdfplumber`, para Primaria;
- el de texto, para ESO y Bachillerato.

Ambos escribirían el mismo formato de ficha.

## 4. Tamaño de la ESO completa

En el Anexo II hay unas 20 materias con título propio: Matemáticas, Lengua, Geografía e Historia, Biología, Física y Química, Inglés, Educación Física, Plástica, Música, Latín y otras, además de las de los ámbitos de diversificación. Al multiplicar por cursos salen unas **60 fichas**, a las que hay que sumar las de Bachillerato. Las 45 fichas que tiene tiza ahora no cubren ni una de las materias troncales de la ESO.

## Incógnitas que quedan abiertas

- Si `herramientas/normativa/comprobar` detecta el fallo de Matemáticas de 1.º. Comprueba que el texto sea literal, y aquí lo es: solo está mal troceado. Si no hay recuentos, seguramente no lo detecta.
- Cómo organizar el repositorio. Hay tres opciones: un repositorio nuevo, un paquete dentro del repositorio de tiza que se publique aparte, o un subdirectorio de Mint Docente. Es una decisión de mantenimiento, no técnica.
- Cómo se instala el paquete en la VM sin red durante el build: una rueda local, como ya se hace con otras herramientas.

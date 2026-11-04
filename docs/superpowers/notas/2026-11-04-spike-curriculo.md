# Spike: fuentes del currículo de Primaria de Madrid (2026-11-04)

## Fuentes (descargadas el 2026-11-04)

| Norma | Boletín | URL | sha256 |
|---|---|---|---|
| Decreto 61/2022, de 13 de julio (ordenación y currículo de Primaria) | BOCM n.º 169, 18-07-2022, 112 págs. | https://www.bocm.es/boletin/CM_Orden_BOCM/2022/07/18/BOCM-20220718-1.PDF | `7087c7d8cdd5a172ecb620a8742acb43864b17563984c407b5f4d498e6399343` |
| Decreto 59/2024, de 12 de junio (modifica el 61/2022) | BOCM n.º 140, 13-06-2024, 5 págs. | https://www.bocm.es/boletin/CM_Orden_BOCM/2024/06/13/BOCM-20240613-2.PDF | `b77c3e24041b2e310c6079d895911c1f9e590c9cf42ede621203568ef43d734a` |
| Orden 130/2023, de 23 de enero (organización, evaluación y autonomía pedagógica en Primaria) | BOCM, 30-01-2023, 60 págs. | https://www.bocm.es/boletin/CM_Orden_BOCM/2023/01/30/BOCM-20230130-23.PDF | `19a649d580692fbd4b187c2c8c93d45925e86e45e77f65dd6219970eb32f6d97` |
| Real Decreto 157/2022, de 1 de marzo (enseñanzas mínimas de Primaria) | BOE-A-2022-3296, 119 págs. | https://www.boe.es/boe/dias/2022/03/02/pdfs/BOE-A-2022-3296.pdf | `21f344161eebd0a298808221c2622074c7e736a76f4bac4e2d695bd080e6750c` |

- **El Decreto 59/2024 no toca los anexos de Primaria.** Solo cambia el art. 9.1: Matemáticas, Lengua y los contenidos de Historia de España de Ciencias Sociales no pueden darse en lengua extranjera. Hay que citarlo como fuente y recogerlo en la skill (programa bilingüe), pero no cambia los datos.
- Consta un recurso contencioso contra el Decreto 61/2022 (P.O. 1367/2022, BOCM del 02-11-2022). No he encontrado su resultado. Queda pendiente comprobar que no se ha anulado nada.
- Un documento de un colegio decía que el Decreto 89/2014 seguía vigente en los cursos pares. Corresponde al calendario de implantación de 2022-23. Desde 2023-24, los seis cursos se rigen por el Decreto 61/2022.

## Estructura del Decreto 61/2022

- **Anexo I:** competencias clave y **descriptores operativos propios de Madrid** al término de Primaria (CCL1…, CP1…, STEM…, CD…, CPSAA…, CC…, CE…, CCEC…). Los textos de los descriptores salen de aquí, no del RD.
- **Anexo II:** una sección por área.
  - **Ciencias de la Naturaleza (CN):** 6 competencias, que corresponden a las CE1–6 de Conocimiento del Medio del RD 157/2022.
  - **Ciencias Sociales (CS):** 5 competencias. **Las CE1–5 de CS son las CE5–9 de Conocimiento del Medio del RD**, con otra numeración. Las CE5–6 del RD se reparten entre CN y CS.
  - **Educación Artística (EA):** un área con competencias y criterios comunes. Los contenidos se dividen en «Bloque I Música y danza» y un bloque de plástica y visual.
  - **Educación Física (EF), Lengua Castellana y Literatura (LCL), Lengua Extranjera: Inglés (ING) y Matemáticas (MAT).**
  - **Educación en Valores Cívicos y Éticos (VCE):** solo en 5º; tiene un único bloque, sin ciclos.
  - **Segunda Lengua Extranjera (SLE) y Tecnología y Robótica (TYR):** son áreas de oferta del centro y tienen **un solo bloque para toda la etapa**, sin ciclos. TYR es propia de Madrid; su texto describe las competencias en el propio decreto.
  - Religión aparece solo en el horario y no tiene currículo en el decreto, así que queda fuera.
- Por ciclo (PRIMER, SEGUNDO y TERCER CICLO), cada área tiene:
  - una tabla de dos columnas, **COMPETENCIAS ESPECÍFICAS | CRITERIOS DE EVALUACIÓN**. Cada fila es una competencia, con su texto, y sus criterios (`n.m`, a veces sin punto final: «2.1 Promover…»);
  - una tabla **CONTENIDOS**, de dos o tres columnas: BLOQUES (A., B.…) | sub-bloque (p. ej. «Iniciación en la actividad científica») | CONOCIMIENTOS, DESTREZAS Y ACTITUDES. Las viñetas usan «-», «x» y «» según el nivel de anidamiento.
- **Madrid los llama «contenidos», no «saberes básicos»,** y **no tienen numeración oficial.**
- **No vincula las competencias con los descriptores.** Ese enlace está en el RD 157/2022: en el BOE aparece 46 veces la frase «se conecta con los siguientes descriptores del Perfil de salida». En CN y CS hay que pasar por la equivalencia de numeración. **TYR y SLE no tienen enlace en el RD.** Queda pendiente comprobar si el texto de TYR en el decreto los menciona; si no, el campo queda vacío.

## Extracción

- `pdftotext` (tanto en modo normal como con `-layout`) mezcla las dos columnas: las competencias se intercalan entre los criterios y los nombres de bloque aparecen en mitad de las viñetas. **No sirve.**
- **`pdfplumber.extract_tables()` funciona.** En las páginas probadas (33, 34 y 37 del PDF), cada fila sale como competencia | criterios, y en los contenidos, bloque | sub-bloque | texto. Las tablas continúan de una página a otra: una fila cuya primera columna está vacía es la continuación de la anterior.
- Cabeceras y pies que hay que eliminar: `BOCM-20220718-1`, `BOCM`, `BOLETÍN OFICIAL DE LA COMUNIDAD DE MADRID`, `B.O.C.M. Núm. 169`, `Pág. N`, `LUNES 18 DE JULIO DE 2022`.
- poppler avisa de una ligadura (`_d_135`). Hay que comprobar en la revisión que no se pierde ningún carácter.

## Orden 130/2023 (para el plan siguiente)

- **Art. 6:**
  - la programación didáctica se hace **por área y ciclo** e incluye unidades de programación por curso, contenidos transversales, métodos, atención a las diferencias y criterios de calificación;
  - cada **unidad de programación** incluye competencias específicas y criterios, contenidos del ciclo, actividades y situaciones de aprendizaje, recursos e instrumentos de evaluación, cada uno con su criterio de calificación.
- Calificaciones IN/SU/BI/NT/SB. Equivalencia: menos de 5 → IN; de 5 a menos de 6 → SU; de 6 a menos de 7 → BI; de 7 a menos de 9 → NT; 9 o más → SB.
- Documentos oficiales de evaluación: actas, expediente, historial, informe final de etapa e informe personal por traslado. Se gestionan en Raíces; aquí solo se redactan borradores.

# Resumen de la revisión del currículo extraído

Referencia: `cache/fuentes/BOCM-20220718-1.PDF` (Decreto 61/2022, BOCM de 18 de julio de 2022), Anexo II para las áreas y Anexo I para los descriptores. Para CN y CS se comprobó además el RD 157/2022 (`cache/fuentes/BOE-A-2022-3296.pdf`).

| Área | Veredicto | N.º de discrepancias | Patrones principales |
|---|---|---:|---|
| CN | APROBADA | 0 | Ninguno |
| CS | NO APROBADA | 1 | El extractor suprime espacios internos en palabras que el boletín parte por justificación: «j udaísmo», «c ristianismo», «i slam» → «judaísmo», «cristianismo», «islam» |
| EA | APROBADA | 0 | Ninguno |
| EF | APROBADA | 0 | Ninguno |
| LCL | APROBADA | 0 | Ninguno |
| ING | APROBADA | 0 | Ninguno |
| MAT | APROBADA | 0 | Ninguno |
| VCE | APROBADA | 0 | Ninguno |
| SLE | APROBADA | 0 | Ninguno |
| TYR | APROBADA | 0 | Ninguno |
| DESCRIPTORES | APROBADA | 0 | Ninguno |

## Dudas que necesitan a una persona

- **CS, contenido 77.º (pág. 35).** El PDF muestra «el j udaísmo, el c ristianismo y el i slam.»; el JSON los escribe unidos. Es la única discrepancia de literalidad detectada. Hay que decidir si se conserva el espacio del boletín (lectura estricta) o se acepta la normalización como corrección tipográfica; la norma de la revisión dice que no se debe corregir.
- **CN.CE6, variante C2 (pág. 23).** «…resolución fomentando respeto, el cuidado…»: faltan la coma y el artículo «el» en el original. El JSON lo reproduce.
- **EF.CE1, variante C3 (pág. 52).** «…activo y adecuado practicando…», sin coma. **EF.CE3, variante C3 (pág. 55).** «…actitudes de, respeto…», con coma tras «de». Reproducidas.
- **LCL.CE9, variante C3 (pág. 67).** «Reflexionar sobre el lenguaje a partir…», sin coma tras «lenguaje». Reproducida. En LCL.CE7, C2 y C3 comparten una variante idéntica que difiere de C1; es correcto que figuren ambos ciclos.
- **ING (págs. 73, 77 y 81).** El original escribe «verb- ing» con espacio tras el guion, y «Wh- questions»; el JSON lo reproduce. «co-presentaciones» está correctamente unido con guion, como en el PDF.
- **MAT (págs. 86 y 88).** El subbloque de C1 dice «Creencias, actitudes valoración personal» (sin «y»); MAT.CE6 C2 dice «…individual y grupal conceptos…» (sin coma); un contenido de la pág. 88 no lleva punto final. Todo se reproduce. «aditivo-multiplicativa» está correctamente unido.
- **SLE (pág. 104).** El contenido de alemán dice «um- Infinitiv», con espacio tras el guion; el JSON lo reproduce.
- **TYR (pág. 110).** El bloque B usa raya: «Mecánica–Ingeniería (Diseño)»; el JSON reproduce la raya.
- **DESCRIPTORES (pág. 31).** CE3 dice «Crea ideas planifica tareas», sin coma; CC1 dice «el valor a la diversidad». El JSON reproduce ambas formas.
- No se ha detectado numeración de criterios repetida ni salteada en ninguna área, ni celdas ilegibles.

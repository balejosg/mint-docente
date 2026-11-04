# Currículo — Ciencias Sociales (CS)

## Fase 1

Fuente: `cache/fuentes/BOCM-20220718-1.PDF`, Anexo II, páginas PDF 28–35 (impresas 42–49).

- Competencias específicas distintas: **5** (CE1–CE5).
- Criterios de evaluación:

| Ciclo | Criterios | Numeración, en orden |
|---|---:|---|
| C1 | 11 | 1.1–1.3; 2.1; 3.1–3.2; 4.1–4.2; 5.1–5.3 |
| C2 | 11 | 1.1–1.3; 2.1; 3.1–3.2; 4.1–4.2; 5.1–5.3 |
| C3 | 11 | 1.1–1.3; 2.1–2.2; 3.1–3.2; 4.1–4.2; 5.1–5.2 |

- Viñetas de contenidos (cada línea con marcador, incluidas las anidadas):

| Ciclo | Total | Nivel 1 («-») | Nivel 2 («•»/x) | Nivel 3 («–»/vacío) |
|---|---:|---:|---:|---:|
| C1 | 21 | 21 | 0 | 0 |
| C2 | 32 | 25 | 2 | 5 |
| C3 | 38 | 31 | 7 | 0 |

### Erratas del boletín

No hay numeración de criterios repetida ni salteada. No se han encontrado celdas ilegibles. Se deja constancia para la fase 2 de que el boletín puede contener erratas tipográficas de redacción que deben reproducirse tal cual.

## Fase 2

ÁREA: CS
VEREDICTO: NO APROBADA
Recuentos PDF vs JSON:
  competencias: 5 vs 5
  C1: criterios 11 vs 11; contenidos 21 vs 21
  C2: criterios 11 vs 11; contenidos 32 vs 32
  C3: criterios 11 vs 11; contenidos 38 vs 38
Discrepancias:
| # | Tipo (falta/sobra/texto/asignación/descriptores/equivalencia/variantes) | Código o ubicación | Página PDF | Dice el PDF | Dice el JSON |
|---|---|---|---|---|---|
| 1 | texto | CS, contenido 77.º (C3; bloque A «Sociedades y territorios»; subbloque «Sociedades en el tiempo») | 35 | «La España de las tres culturas: el j udaísmo, el c ristianismo y el i slam.» | «La España de las tres culturas: el judaísmo, el cristianismo y el islam.» |

Patrones:
- El extractor elimina los espacios internos que el boletín deja al partir palabras por justificación: «j udaísmo», «c ristianismo» e «i slam» se convierten en «judaísmo», «cristianismo» e «islam». Es un fallo de literalidad: el espacio forma parte de la línea del boletín y no debe suprimirse.

Dudas que necesitan a una persona:
- En CS.CE1, variante C3 (pág. 33), el boletín dice «del medio analizando su organización…», sin coma tras «medio»; el JSON lo reproduce fielmente.
- Los descriptores de CS.CE1–CE5 coinciden con los del RD 157/2022 para CMNSC.CE5–CE9, y la equivalencia declarada en la introducción del área es la correcta.
- No hay numeración repetida ni salteada, ni celdas ilegibles. La separación de palabras de la página 35 es la discrepancia 1, no una errata que el JSON haya reproducido.
- El hueco de «j udaísmo», «c ristianismo» e «i slam» es de unos 1,1 pt (menor que un espacio normal de la línea, pero muy superior al kerning ordinario). Si una persona lo considera una simple diferencia de espaciado y no una palabra partida, podría reclasificarse; con la regla literal de esta revisión se mantiene como discrepancia.

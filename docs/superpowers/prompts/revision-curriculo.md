# Prompt: revisión independiente del currículo extraído (un área por sesión)

**Cómo usarlo:**
1. Abre una **sesión nueva por área** en un LLM que lea PDF: CN, CS, EA, EF, LCL, ING, MAT, VCE, SLE y TYR. Los descriptores se revisan en una sesión aparte, con la variante del final.
2. **Mensaje 1:** adjunta el PDF del Decreto 61/2022 (`BOCM-20220718-1.PDF`) y pega el prompt de la fase 1, sustituyendo `{ÁREA}` y `{NOMBRE DEL ÁREA}`. **No adjuntes todavía el JSON.**
3. **Mensaje 2:** cuando haya respondido, pega el prompt de la fase 2 y adjunta el fragmento del JSON de esa área. Se obtiene con:
   `jq '.areas[] | select(.codigo=="{ÁREA}")' files/curriculo/primaria-madrid.json > {ÁREA}.json`
   En CN y CS adjunta también el PDF del RD 157/2022 (`BOE-A-2022-3296.pdf`).
4. Guarda la respuesta final en `docs/superpowers/revisiones/curriculo-{ÁREA}.md`. Copia los recuentos de la fase 1 en `files/curriculo/recuentos-esperados.json`.

---

## Fase 1 (solo el PDF)

Eres revisor de normativa educativa. Te adjunto el **Decreto 61/2022, de 13 de julio**, de la Comunidad de Madrid (BOCM n.º 169, 18-07-2022), que establece el currículo de Educación Primaria.

Trabaja **solo** con el **Anexo II**, apartado del área **«{NOMBRE DEL ÁREA}»**, desde su título hasta el título del área siguiente. Lee las tablas tal como están maquetadas: en la columna izquierda están las competencias específicas y en la derecha los criterios de evaluación. Las tablas de CONTENIDOS tienen columnas de bloque, de sub-bloque (a veces) y de «Conocimientos, destrezas y actitudes». Las tablas continúan de una página a otra.

Sin ningún otro material, cuenta y responde **exactamente** con este formato:

```
ÁREA: {ÁREA}
Páginas del PDF que abarca: <primera>–<última>
Organización: <por ciclos (PRIMER/SEGUNDO/TERCER CICLO) | un único bloque para toda la etapa o curso>
Competencias específicas distintas en el área: <n>
Por ciclo (o «E» si es única):
  <ciclo>: criterios=<n>  viñetas de contenidos=<n>
Lista de criterios por ciclo (solo la numeración, en orden):
  <ciclo>: 1.1, 1.2, 2.1, …
Dudas de lectura: <celdas ilegibles, numeración salteada o repetida en el original, erratas del propio boletín, o «ninguna»>
```

Una «viñeta de contenidos» es cada línea que empieza por un marcador («-», «x», «», «•» u otro), incluidas las anidadas. Cuenta cada viñeta una vez, aunque ocupe varias líneas.

No resumas ni interpretes: es un recuento. Si el propio boletín tiene una errata, por ejemplo un criterio numerado dos veces, anótala en «Dudas de lectura» tal como aparece.

---

## Fase 2 (con el JSON del área)

Ahora te adjunto `{ÁREA}.json`: lo que un programa ha extraído automáticamente de ese mismo apartado del PDF. Su estructura es:
- `competencias[]`: `{codigo, numero, texto, descriptores[], equivalencia_rd}`;
- `criterios[]`: `{codigo, ciclo, competencia, texto, pagina}`. El código es `{ÁREA}.<n>.<m>.<ciclo>`, con `ciclo` = C1, C2, C3 o E;
- `contenidos[]`: `{ciclo, bloque, subbloque, nivel, texto, pagina}`.

Compáralo con el PDF **elemento a elemento**, sin muestrear. Comprueba:

1. **Que no falte nada ni sobre nada:** cada criterio y cada viñeta de contenidos del PDF está en el JSON exactamente una vez, y no hay nada en el JSON que no esté en el PDF. Usa tus recuentos de la fase 1.
2. **Textos literales:** el texto coincide con el del boletín palabra por palabra. Se admiten diferencias de salto de línea y de espacios; **no** se admiten palabras cortadas o unidas («investi gación», «delmedio»), tildes o letras perdidas, frases de otra celda mezcladas, ni restos de cabecera o pie de página («BOCM», «Pág.», la fecha).
3. **Asignación correcta:**
   - cada criterio cuelga de su competencia y del ciclo correcto;
   - cada contenido, de su bloque, su sub-bloque y su ciclo;
   - el `nivel` de anidamiento de las viñetas es correcto.
4. **Competencias:** el texto completo de cada una está en `texto`, sin partes de criterios mezcladas.
5. **Solo en CN y CS:**
   - `equivalencia_rd` coincide con la correspondencia que declara el propio decreto en la introducción del área: CN CE1–6 = Conocimiento del Medio CE1–6; CS CE1–5 = Conocimiento del Medio CE5–9;
   - los `descriptores` de cada competencia coinciden con los que da el RD 157/2022, Anexo II, área de Conocimiento del Medio Natural, Social y Cultural, en la frase «se conecta con los siguientes descriptores del Perfil de salida».
6. **En el resto de áreas:** los `descriptores` de cada competencia coinciden con los que da el RD 157/2022 para la misma área y competencia. En TYR y SLE, que no existen en el RD, el campo debe estar vacío salvo que el propio decreto los enumere.

Responde **exactamente** con este formato:

```
ÁREA: {ÁREA}
VEREDICTO: APROBADA | NO APROBADA
Recuentos PDF vs JSON:
  competencias: <pdf> vs <json>
  <ciclo>: criterios <pdf> vs <json>; contenidos <pdf> vs <json>
Discrepancias (una fila por problema; «ninguna» si no hay):
| # | Tipo (falta/sobra/texto/asignación/descriptores/equivalencia) | Código o ubicación | Página PDF | Dice el PDF | Dice el JSON |
Patrones (si un mismo error se repite, descríbelo una vez para corregir el extractor):
Dudas que necesitan a una persona:
```

Solo se aprueba con **cero discrepancias**. Una errata del propio boletín que el JSON reproduce fielmente no cuenta como discrepancia; anótala en «Dudas». No propongas correcciones al texto oficial ni lo mejores: la referencia es el boletín, incluidos sus errores.

---

## Variante: descriptores (sesión aparte)

Adjunta el PDF del decreto y el resultado de `jq '.descriptores' files/curriculo/primaria-madrid.json`. En lugar de las dos fases anteriores, pega esto:

> Te adjunto el Decreto 61/2022 de Madrid y un JSON con los descriptores operativos que un programa ha extraído de su **Anexo I** («descriptores operativos… al término de la etapa de Educación Primaria en la Comunidad de Madrid»). Primero cuenta, en el PDF, cuántos descriptores hay por competencia clave (CCL, CP, STEM, CD, CPSAA, CC, CE y CCEC) y lista sus códigos. Después compara el JSON elemento a elemento: códigos, competencia clave y texto literal, con los mismos criterios de literalidad (sin palabras cortadas, sin letras perdidas, sin restos de cabecera). Responde con el formato de la fase 2 y pon «DESCRIPTORES» como área.

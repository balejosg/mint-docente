#!/usr/bin/env python3
"""Extrae el currículo de Primaria de la Comunidad de Madrid a JSON.

Solo se usa en el anfitrión; no entra en la imagen. Uso:

    .venv-dev/bin/python -I herramientas/extraer-curriculo.py \
        --fuentes cache/fuentes --salida files/curriculo/primaria-madrid.json

- Descarga en --fuentes las normas que falten y comprueba su sha256.
- Anexo I del Decreto 61/2022: descriptores operativos.
- Anexo II del Decreto 61/2022: competencias (con sus «variantes» de redacción
  por ciclo), criterios y contenidos por área y ciclo, leyendo las tablas con
  pdfplumber.
- RD 157/2022: enlace competencia → descriptores (SLE y TYR lo traen en el
  propio decreto).

Los avisos van a la salida de error: hay que leerlos antes de dar el JSON por bueno.
"""
import argparse
import datetime
import hashlib
import json
import re
import sys
import urllib.request
from pathlib import Path

import pdfplumber

FUENTES = [
    {
        "norma": "Decreto 61/2022, de 13 de julio, del Consejo de Gobierno, por el que se "
        "establecen para la Comunidad de Madrid la ordenación y el currículo de la etapa "
        "de Educación Primaria",
        "boletin": "BOCM n.º 169",
        "fecha": "2022-07-18",
        "url": "https://www.bocm.es/boletin/CM_Orden_BOCM/2022/07/18/BOCM-20220718-1.PDF",
        "sha256": "7087c7d8cdd5a172ecb620a8742acb43864b17563984c407b5f4d498e6399343",
    },
    {
        "norma": "Decreto 59/2024, de 12 de junio, del Consejo de Gobierno, por el que se "
        "modifica el Decreto 61/2022, de 13 de julio",
        "boletin": "BOCM n.º 140",
        "fecha": "2024-06-13",
        "url": "https://www.bocm.es/boletin/CM_Orden_BOCM/2024/06/13/BOCM-20240613-2.PDF",
        "sha256": "b77c3e24041b2e310c6079d895911c1f9e590c9cf42ede621203568ef43d734a",
    },
    {
        "norma": "Orden 130/2023, de 23 de enero, de la Vicepresidencia, Consejería de "
        "Educación y Universidades, por la que se regulan aspectos de organización, "
        "funcionamiento, evaluación y autonomía pedagógica de la Educación Primaria",
        "boletin": "BOCM",
        "fecha": "2023-01-30",
        "url": "https://www.bocm.es/boletin/CM_Orden_BOCM/2023/01/30/BOCM-20230130-23.PDF",
        "sha256": "19a649d580692fbd4b187c2c8c93d45925e86e45e77f65dd6219970eb32f6d97",
    },
    {
        "norma": "Real Decreto 157/2022, de 1 de marzo, por el que se establecen la "
        "ordenación y las enseñanzas mínimas de la Educación Primaria",
        "boletin": "BOE-A-2022-3296",
        "fecha": "2022-03-02",
        "url": "https://www.boe.es/boe/dias/2022/03/02/pdfs/BOE-A-2022-3296.pdf",
        "sha256": "21f344161eebd0a298808221c2622074c7e736a76f4bac4e2d695bd080e6750c",
    },
]
DECRETO, RD = FUENTES[0], FUENTES[3]

# Áreas del Anexo II, en el orden del decreto. «rd» es el título del área en el
# RD 157/2022 y «desfase_rd» lo que se suma al número de competencia para dar la
# del RD (CS.CE1 = CMNSC.CE5). Los alias responden a los nombres coloquiales.
AREAS = [
    {"codigo": "CN", "nombre": "Ciencias de la Naturaleza", "ciclos": ["C1", "C2", "C3"],
     "rd": "Conocimiento del Medio Natural, Social y Cultural", "prefijo_rd": "CMNSC", "desfase_rd": 0,
     "alias": ["ciencias de la naturaleza", "naturales", "ciencias naturales", "naturaleza",
               "conocimiento del medio", "ciencias"]},
    {"codigo": "CS", "nombre": "Ciencias Sociales", "ciclos": ["C1", "C2", "C3"],
     "rd": "Conocimiento del Medio Natural, Social y Cultural", "prefijo_rd": "CMNSC", "desfase_rd": 4,
     "alias": ["ciencias sociales", "sociales", "conocimiento del medio", "ciencias"]},
    {"codigo": "EA", "nombre": "Educación Artística", "ciclos": ["C1", "C2", "C3"],
     "rd": "Educación Artística",
     "alias": ["educación artística", "artística", "plástica", "música", "educación plástica",
               "educación plástica y visual", "educación musical", "arte"]},
    {"codigo": "EF", "nombre": "Educación Física", "ciclos": ["C1", "C2", "C3"],
     "rd": "Educación Física",
     "alias": ["educación física", "gimnasia", "física"]},
    {"codigo": "LCL", "nombre": "Lengua Castellana y Literatura", "ciclos": ["C1", "C2", "C3"],
     "rd": "Lengua Castellana y Literatura",
     "alias": ["lengua castellana y literatura", "lengua", "lengua castellana", "castellano",
               "literatura"]},
    {"codigo": "ING", "nombre": "Lengua Extranjera: Inglés", "ciclos": ["C1", "C2", "C3"],
     "rd": "Lengua Extranjera",
     "alias": ["lengua extranjera: inglés", "inglés", "lengua extranjera", "idioma"]},
    {"codigo": "MAT", "nombre": "Matemáticas", "ciclos": ["C1", "C2", "C3"],
     "rd": "Matemáticas",
     "alias": ["matemáticas", "mates", "mate"]},
    {"codigo": "VCE", "nombre": "Educación en Valores Cívicos y Éticos", "ciclos": ["C3"],
     "rd": "Educación en Valores Cívicos y Éticos",
     "alias": ["educación en valores cívicos y éticos", "valores", "valores cívicos",
               "valores cívicos y éticos", "educación en valores"]},
    {"codigo": "SLE", "nombre": "Segunda Lengua Extranjera", "ciclos": ["E"], "rd": None,
     "alias": ["segunda lengua extranjera", "segunda lengua", "francés", "alemán", "italiano",
               "portugués", "idioma"]},
    {"codigo": "TYR", "nombre": "Tecnología y Robótica", "ciclos": ["E"], "rd": None,
     "alias": ["tecnología y robótica", "robótica", "tecnología", "programación"]},
]

CICLOS = {"PRIMER CICLO": "C1", "SEGUNDO CICLO": "C2", "TERCER CICLO": "C3"}
RE_DESCRIPTOR = re.compile(r"^(CCL|CP|STEM|CD|CPSAA|CC|CE|CCEC)([0-9])\.\s+(.*)$")
RE_COMPETENCIA = re.compile(r"^(\d+)\.\s+(.*)$", re.S)
RE_CRITERIO = re.compile(r"^(\d+)\.(\d{1,2})\.?\s+(.*)$")
VINETAS = [("- ", 1), ("(cid:2) ", 2), ("(cid:3) ", 3)]  # «-», «•» y «–» en el PDF
RE_LISTA_DESCRIPTORES = r"siguientes descriptores del [Pp]erfil de\s+salida:\s*([A-Z0-9,\s]+?)\."

# Recortes que dejan fuera cabeceras, pies y el texto vertical del margen.
RECORTE_BOCM = (0, 95, 560, None)
RECORTE_BOE = (0, 100, 560, None)

avisos = []


def aviso(msg):
    avisos.append(msg)
    print(f"AVISO: {msg}", file=sys.stderr)


# --- Fuentes -----------------------------------------------------------------

def sha256(ruta):
    h = hashlib.sha256()
    with open(ruta, "rb") as f:
        for bloque in iter(lambda: f.read(1 << 20), b""):
            h.update(bloque)
    return h.hexdigest()


def preparar_fuentes(dir_fuentes):
    """Descarga las normas que falten y comprueba el sha256 de todas."""
    dir_fuentes.mkdir(parents=True, exist_ok=True)
    rutas = {}
    for f in FUENTES:
        ruta = dir_fuentes / f["url"].rsplit("/", 1)[1]
        if not ruta.exists():
            print(f"Descargando {f['url']}", file=sys.stderr)
            peticion = urllib.request.Request(f["url"], headers={"User-Agent": "Mozilla/5.0"})
            temporal = ruta.with_suffix(ruta.suffix + ".parcial")
            with urllib.request.urlopen(peticion, timeout=120) as r, open(temporal, "wb") as s:
                s.write(r.read())
            temporal.rename(ruta)
        real = sha256(ruta)
        if real != f["sha256"]:
            sys.exit(f"ERROR: el sha256 de {ruta} no coincide con el de la nota del spike "
                     f"({real} ≠ {f['sha256']}). Se aborta.")
        rutas[f["url"]] = ruta
    return rutas


# --- Utilidades de texto -------------------------------------------------------

def unir(a, b):
    """Une dos líneas de un mismo párrafo. Un guion al final de línea pegado a
    una letra es parte de un compuesto («sintáctico-discursivos»): se conserva y
    no se pone espacio."""
    a, b = a.rstrip(), b.strip()
    if not a:
        return b
    if not b:
        return a
    if a[-1] in "-–" and len(a) > 1 and a[-2].isalpha() and b[0].isalpha():
        return a + b
    return a + " " + b


def unir_lineas(texto):
    resultado = ""
    for linea in texto.split("\n"):
        resultado = unir(resultado, linea)
    return re.sub(r" {2,}", " ", resultado).strip()


def recortar(pagina, recorte):
    x0, top, x1, bottom = recorte
    return pagina.crop((x0, top, x1, bottom or pagina.height))


def contenida(a, b):
    return a != b and a[0] >= b[0] - 1 and a[1] >= b[1] - 1 and a[2] <= b[2] + 1 and a[3] <= b[3] + 1


# --- Anexo I: descriptores -----------------------------------------------------

def extraer_descriptores(pdf):
    lineas = []
    dentro = False
    for pagina in pdf.pages:
        texto = recortar(pagina, RECORTE_BOCM).extract_text() or ""
        for linea in texto.split("\n"):
            linea = linea.strip()
            if linea == "ANEXO I":
                dentro = True
                continue
            if linea == "ANEXO II":
                dentro = False
                break
            if dentro:
                lineas.append(linea)
        if lineas and not dentro:
            break
    descriptores = []
    for linea in lineas:
        m = RE_DESCRIPTOR.match(linea)
        if m:
            descriptores.append({"codigo": m[1] + m[2], "competencia_clave": m[1], "texto": m[3]})
        elif linea == "Descriptores operativos" or re.match(r"^[a-h]\) Competencia", linea):
            continue
        elif descriptores:
            descriptores[-1]["texto"] = unir(descriptores[-1]["texto"], linea)
    return descriptores


# --- RD 157/2022: competencia → descriptores --------------------------------------

def extraer_descriptores_rd(pdf):
    """{título del área en el RD: [[descriptores de CE1], [de CE2], …]}"""
    lineas = []
    for pagina in pdf.pages:
        lineas += (recortar(pagina, RECORTE_BOE).extract_text() or "").split("\n")
    lineas = [l.strip() for l in lineas]
    inicio = lineas.index("ANEXO II")
    titulos = sorted({a["rd"] for a in AREAS if a["rd"]})
    posiciones = sorted((lineas.index(t, inicio), t) for t in titulos)
    resultado = {}
    for i, (pos, titulo) in enumerate(posiciones):
        fin = posiciones[i + 1][0] if i + 1 < len(posiciones) else len(lineas)
        texto = " ".join(lineas[pos:fin])
        listas = re.findall(r"se conecta con los siguientes descriptores del Perfil de\s+salida:"
                            r"\s*([A-Z0-9,\s]+?)\.", texto)
        resultado[titulo] = [[d.strip() for d in l.split(",") if d.strip()] for l in listas]
    return resultado


# --- Anexo II: lectura de páginas en eventos ------------------------------------

def eventos_anexo_ii(pdf):
    """Recorre el Anexo II y produce eventos ordenados: títulos de área, cabeceras
    de ciclo y filas de tabla con la posición de cada celda."""
    titulos = {a["nombre"]: a["codigo"] for a in AREAS}
    dentro = False
    for n, pagina_entera in enumerate(pdf.pages, 1):
        pagina = recortar(pagina_entera, RECORTE_BOCM)
        texto = pagina.extract_text() or ""
        if not dentro:
            if "ANEXO II" in texto.split("\n"):
                dentro = True
            else:
                continue
        if "ANEXO III" in texto.split("\n"):
            break
        tablas = pagina.find_tables()
        tablas = [t for t in tablas if not any(contenida(t.bbox, o.bbox) for o in tablas)]

        def en_tabla(o):
            return any(t.bbox[0] - 1 <= o["x0"] and o["x1"] <= t.bbox[2] + 1 and
                       t.bbox[1] - 1 <= o["top"] and o["bottom"] <= t.bbox[3] + 1 for t in tablas)

        fuera = pagina.filter(lambda o: not (o.get("object_type") == "char" and en_tabla(o)))
        eventos = []
        centro = pagina_entera.width / 2
        for l in fuera.extract_text_lines():
            t = l["text"].strip()
            if t in titulos and abs((l["x0"] + l["x1"]) / 2 - centro) < 30:
                eventos.append((l["top"], 0, ("area", titulos[t])))
            elif t in CICLOS:
                eventos.append((l["top"], 0, ("ciclo", CICLOS[t])))
        for tabla in tablas:
            # pdfplumber a veces da una celda grande y, además, una fila por cada
            # línea que contiene: se descartan las celdas dentro de otra ya leída.
            leidas = []
            for fila, datos in zip(tabla.rows, tabla.extract()):
                celdas = []
                for c, d in zip(fila.cells, datos):
                    if c is None or c[2] - c[0] < 12:
                        continue
                    if any(contenida(c, l) for l in leidas):
                        continue
                    leidas.append(c)
                    celdas.append((c[0], c[2], (d or "").strip()))
                if celdas:
                    eventos.append((fila.bbox[1], 1, ("fila", tabla.bbox[0], sorted(celdas))))
        for _, _, ev in sorted(eventos, key=lambda e: (e[0], e[1])):
            yield n, ev


# --- Anexo II: interpretación de las tablas --------------------------------------

class Area:
    def __init__(self, definicion):
        self.d = definicion
        self.codigo = definicion["codigo"]
        self.ciclo = definicion["ciclos"][0] if len(definicion["ciclos"]) == 1 else None
        self.seccion = None          # None (introducción), "competencias" o "contenidos"
        self.corte_criterios = None  # x a partir de la cual una celda es de criterios
        self.corte_contenidos = None
        self.competencias = {}       # ciclo -> {numero: texto}
        self.criterios = []
        self.contenidos = []
        self.comp_actual = None      # número de la competencia de la fila en curso
        self.ultimo = None           # último texto en curso: criterio o contenido
        self.bloque = None
        self.subbloque = None
        self.gran_bloque = None      # «BLOQUE I Música y danza» (Educación Artística)
        self.gran_bloque_nuevo = None

    # cabeceras y marcas
    def nuevo_ciclo(self, ciclo):
        if len(self.d["ciclos"]) == 1:
            aviso(f"{self.codigo}: cabecera de ciclo {ciclo} en un área de ciclo único")
            return
        self.ciclo, self.seccion = ciclo, None
        self.bloque = self.subbloque = self.gran_bloque = self.ultimo = None

    def fila(self, pagina, x_tabla, celdas):
        textos = [t for _, _, t in celdas if t]
        if not textos:
            return
        if len(textos) == 1 and textos[0] in CICLOS:
            return self.nuevo_ciclo(CICLOS[textos[0]])
        if "CRITERIOS DE EVALUACIÓN" in textos:
            x = next(x0 for x0, _, t in celdas if t == "CRITERIOS DE EVALUACIÓN")
            self.corte_criterios = x - 12
            if self.ciclo is None:
                aviso(f"{self.codigo}: competencias sin cabecera de ciclo (pág. {pagina})")
            if self.seccion != "competencias":
                self.ultimo = None
            self.seccion = "competencias"
            return
        if textos == ["CONTENIDOS"]:
            self.seccion, self.ultimo = "contenidos", None
            self.bloque = self.subbloque = self.gran_bloque = None
            return
        if self.seccion is None:
            return  # introducción del área maquetada como tabla
        if "CONOCIMIENTOS, DESTREZAS Y ACTITUDES" in textos:
            x = next(x0 for x0, _, t in celdas if t == "CONOCIMIENTOS, DESTREZAS Y ACTITUDES")
            self.corte_contenidos = x - 12
            resto = [t for t in textos if t not in ("CONOCIMIENTOS, DESTREZAS Y ACTITUDES", "BLOQUES")]
            if resto and re.match(r"^BLOQUE [IVX]+\b", resto[0]):
                self.gran_bloque_nuevo = unir_lineas(resto[0])
            elif resto:
                aviso(f"{self.codigo}: cabecera de contenidos con texto inesperado {resto} (pág. {pagina})")
            return
        if self.seccion == "competencias":
            self.fila_competencias(pagina, celdas)
        else:
            self.fila_contenidos(pagina, x_tabla, celdas)

    # competencias y criterios
    def fila_competencias(self, pagina, celdas):
        izquierda = [t for x0, _, t in celdas if x0 < self.corte_criterios and t]
        derecha = [t for x0, _, t in celdas if x0 >= self.corte_criterios and t]
        if len(izquierda) > 1 or len(derecha) > 1:
            aviso(f"{self.codigo}: fila de competencias con varias celdas por columna (pág. {pagina})")
        comps = self.competencias.setdefault(self.ciclo, {})
        if izquierda:
            texto = "\n".join(izquierda)
            m = RE_COMPETENCIA.match(texto)
            if m:
                self.comp_actual = int(m[1])
                if self.comp_actual in comps:
                    aviso(f"{self.codigo}: competencia {m[1]} repetida en {self.ciclo} (pág. {pagina})")
                comps[self.comp_actual] = unir_lineas(m[2])
            elif self.comp_actual is not None:
                comps[self.comp_actual] = unir(comps[self.comp_actual], unir_lineas(texto))
            else:
                aviso(f"{self.codigo}: texto de competencia sin número (pág. {pagina}): {texto[:60]!r}")
        for linea in "\n".join(derecha).split("\n") if derecha else []:
            m = RE_CRITERIO.match(linea.strip())
            if m:
                n, num = int(m[1]), int(m[2])
                if n != self.comp_actual:
                    aviso(f"{self.codigo}: criterio {n}.{num} bajo la competencia {self.comp_actual} "
                          f"en {self.ciclo} (pág. {pagina})")
                self.ultimo = {"n": n, "m": num, "ciclo": self.ciclo, "texto": m[3].strip(),
                               "pagina": pagina}
                self.criterios.append(self.ultimo)
            elif self.ultimo is not None and "m" in self.ultimo:
                self.ultimo["texto"] = unir(self.ultimo["texto"], linea)
            else:
                aviso(f"{self.codigo}: texto de criterio sin numeración (pág. {pagina}): {linea[:60]!r}")

    # contenidos
    def fila_contenidos(self, pagina, x_tabla, celdas):
        if self.corte_contenidos is None:
            aviso(f"{self.codigo}: contenidos sin cabecera de columnas (pág. {pagina})")
            return
        texto = [t for x0, _, t in celdas if x0 >= self.corte_contenidos]
        izquierda = [(x0, t) for x0, _, t in celdas if x0 < self.corte_contenidos]
        col_bloque = [t for x0, t in izquierda if x0 <= x_tabla + 10]
        col_sub = [t for x0, t in izquierda if x0 > x_tabla + 10]
        if len(texto) > 1 or len(col_bloque) > 1 or len(col_sub) > 1:
            aviso(f"{self.codigo}: fila de contenidos con varias celdas por columna (pág. {pagina})")
        texto = "\n".join(texto)
        bloque = unir_lineas("\n".join(col_bloque)) or None
        sub = unir_lineas("\n".join(col_sub)) or None
        lineas = [l.strip() for l in texto.split("\n") if l.strip()]
        continua = bool(lineas) and not any(lineas[0].startswith(v) for v, _ in VINETAS)

        # Fila sin contenidos: continuación del título de un «BLOQUE I/II» de Artística.
        if not lineas:
            if self.gran_bloque_nuevo is not None and (bloque or sub):
                self.gran_bloque_nuevo = unir(self.gran_bloque_nuevo, bloque or sub)
            elif bloque or sub:
                aviso(f"{self.codigo}: celda de bloque sin contenidos (pág. {pagina}): {bloque or sub!r}")
            return
        if self.gran_bloque_nuevo is not None:
            if self.gran_bloque_nuevo != self.gran_bloque:
                self.gran_bloque, self.subbloque = self.gran_bloque_nuevo, None
            self.gran_bloque_nuevo = None

        # En Artística, la columna izquierda es el sub-bloque del «BLOQUE I/II».
        if self.gran_bloque is not None:
            bloque, sub = None, bloque if bloque else sub
        if bloque:
            if continua and bloque[0].islower():
                self.renombrar("bloque", unir(self.bloque, bloque))
            else:
                self.bloque, self.subbloque = bloque, None
        if sub:
            if continua and sub[0].islower():
                self.renombrar("subbloque", unir(self.subbloque, sub))
            else:
                self.subbloque = sub
        if self.gran_bloque is not None:
            self.bloque = self.gran_bloque
        if self.bloque is None:
            aviso(f"{self.codigo}: contenidos sin bloque (pág. {pagina})")
        if (bloque or (sub and not sub[0].islower())) and continua:
            aviso(f"{self.codigo}: bloque o sub-bloque nuevo que empieza sin viñeta (pág. {pagina}): "
                  f"{lineas[0][:60]!r}")

        for linea in lineas:
            nivel = next((nv for v, nv in VINETAS if linea.startswith(v)), None)
            if nivel is not None:
                marca = next(v for v, nv in VINETAS if nv == nivel)
                self.ultimo = {"ciclo": self.ciclo, "bloque": self.bloque, "subbloque": self.subbloque,
                               "nivel": nivel, "texto": linea[len(marca):].strip(), "pagina": pagina}
                self.contenidos.append(self.ultimo)
            elif self.ultimo is not None and "nivel" in self.ultimo:
                self.ultimo["texto"] = unir(self.ultimo["texto"], linea)
            else:
                aviso(f"{self.codigo}: contenido sin viñeta y sin anterior (pág. {pagina}): {linea[:60]!r}")

    def renombrar(self, campo, nuevo):
        """Un nombre de bloque o sub-bloque partido entre dos páginas."""
        viejo = getattr(self, campo)
        for c in self.contenidos:
            if c["ciclo"] == self.ciclo and c[campo] == viejo and (campo == "bloque" or c["bloque"] == self.bloque):
                c[campo] = nuevo
        setattr(self, campo, nuevo)

    # resultado
    def resultado(self, descriptores_rd, descriptores_decreto):
        codigo = self.codigo
        textos = {}
        for ciclo in self.d["ciclos"]:
            comps = self.competencias.get(ciclo, {})
            if not comps:
                aviso(f"{codigo}: no hay competencias en {ciclo}")
            for n, t in comps.items():
                textos.setdefault(n, {}).setdefault(t, []).append(ciclo)
        competencias = []
        for n in sorted(textos):
            # El decreto repite las competencias en cada ciclo y a veces cambia la
            # redacción (comas, alguna palabra, erratas). «texto» es la del primer
            # ciclo en que aparece y «variantes» guarda, por ciclo, las que difieren.
            por_ciclo = {c: t for t, ciclos in textos[n].items() for c in ciclos}
            presentes = [c for c in self.d["ciclos"] if c in por_ciclo]
            if presentes != self.d["ciclos"]:
                aviso(f"{codigo}.CE{n}: solo aparece en {presentes}")
            texto = por_ciclo[presentes[0]]
            variantes = {c: t for c, t in por_ciclo.items() if t != texto}
            if variantes:
                print(f"Nota: {codigo}.CE{n} cambia de redacción en {sorted(variantes)}", file=sys.stderr)
            if self.d.get("rd"):
                n_rd = n + self.d.get("desfase_rd", 0)
                listas = descriptores_rd[self.d["rd"]]
                desc = listas[n_rd - 1] if n_rd <= len(listas) else []
                if not desc:
                    aviso(f"{codigo}.CE{n}: sin descriptores en el RD")
            else:
                desc = descriptores_decreto[n - 1] if n <= len(descriptores_decreto) else []
            competencias.append({
                "codigo": f"{codigo}.CE{n}", "numero": n, "texto": texto,
                "variantes": dict(sorted(variantes.items())), "descriptores": desc,
                "equivalencia_rd": (f"{self.d['prefijo_rd']}.CE{n + self.d['desfase_rd']}"
                                    if "prefijo_rd" in self.d else None),
            })
        if self.d.get("rd") is None and len(descriptores_decreto) != len(competencias):
            aviso(f"{codigo}: {len(descriptores_decreto)} listas de descriptores en el decreto para "
                  f"{len(competencias)} competencias")

        criterios, vistos = [], set()
        anterior = {}
        for c in self.criterios:
            cod = f"{codigo}.{c['n']}.{c['m']}.{c['ciclo']}"
            if cod in vistos:
                aviso(f"{cod}: código repetido (pág. {c['pagina']})")
            vistos.add(cod)
            prev = anterior.get(c["ciclo"])
            esperado = [(prev[0], prev[1] + 1), (prev[0] + 1, 1)] if prev else [(1, 1)]
            if (c["n"], c["m"]) not in esperado:
                aviso(f"{cod}: numeración no consecutiva tras {prev} (pág. {c['pagina']})")
            anterior[c["ciclo"]] = (c["n"], c["m"])
            criterios.append({"codigo": cod, "ciclo": c["ciclo"], "competencia": f"{codigo}.CE{c['n']}",
                              "texto": c["texto"], "pagina": c["pagina"]})
        return {
            "codigo": codigo, "nombre": self.d["nombre"], "alias": self.d["alias"],
            "ciclos": self.d["ciclos"], "competencias": competencias, "criterios": criterios,
            "contenidos": self.contenidos,
        }


def descriptores_en_decreto(pdf, codigo):
    """SLE y TYR enumeran sus descriptores en el propio decreto, en la
    «Descripción de las competencias específicas del área»."""
    nombre = next(a["nombre"] for a in AREAS if a["codigo"] == codigo)
    siguiente = {a["codigo"]: AREAS[i + 1]["nombre"] if i + 1 < len(AREAS) else "ANEXO III"
                 for i, a in enumerate(AREAS)}[codigo]
    texto = []
    dentro = False
    for pagina in pdf.pages:
        for linea in (recortar(pagina, RECORTE_BOCM).extract_text() or "").split("\n"):
            linea = linea.strip()
            if linea == nombre:
                dentro = True
            elif linea == siguiente and dentro:
                dentro = False
            elif dentro:
                texto.append(linea)
    listas = re.findall(RE_LISTA_DESCRIPTORES, " ".join(texto))
    return [[d.strip() for d in l.split(",") if d.strip()] for l in listas]


def extraer_anexo_ii(pdf, descriptores_rd):
    areas, actual = [], None
    definiciones = {a["codigo"]: a for a in AREAS}
    for pagina, ev in eventos_anexo_ii(pdf):
        if ev[0] == "area":
            actual = Area(definiciones[ev[1]])
            areas.append(actual)
        elif actual is None:
            continue
        elif ev[0] == "ciclo":
            actual.nuevo_ciclo(ev[1])
        else:
            actual.fila(pagina, ev[1], ev[2])
    if [a.codigo for a in areas] != [a["codigo"] for a in AREAS]:
        aviso(f"áreas encontradas: {[a.codigo for a in areas]}")
    resultado = []
    for a in areas:
        decreto = descriptores_en_decreto(pdf, a.codigo) if a.d["rd"] is None else []
        resultado.append(a.resultado(descriptores_rd, decreto))
    return resultado


def main():
    p = argparse.ArgumentParser(description="Extrae el currículo de Primaria de Madrid a JSON.")
    p.add_argument("--fuentes", type=Path, required=True, help="carpeta de los PDF (se descargan si faltan)")
    p.add_argument("--salida", type=Path, required=True, help="archivo JSON de salida")
    args = p.parse_args()

    rutas = preparar_fuentes(args.fuentes)
    with pdfplumber.open(rutas[RD["url"]]) as pdf:
        descriptores_rd = extraer_descriptores_rd(pdf)
    for a in AREAS:
        if a["rd"] and not descriptores_rd.get(a["rd"]):
            aviso(f"RD: no hay descriptores para «{a['rd']}»")
    with pdfplumber.open(rutas[DECRETO["url"]]) as pdf:
        descriptores = extraer_descriptores(pdf)
        areas = extraer_anexo_ii(pdf, descriptores_rd)

    datos = {
        "version": datetime.date.today().isoformat(),
        "fuentes": [{k: f[k] for k in ("norma", "boletin", "fecha", "url", "sha256")} for f in FUENTES],
        "descriptores": descriptores,
        "areas": areas,
    }
    args.salida.parent.mkdir(parents=True, exist_ok=True)
    with open(args.salida, "w", encoding="utf-8") as f:
        json.dump(datos, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"Escrito {args.salida}: {len(descriptores)} descriptores, {len(areas)} áreas, "
          f"{len(avisos)} avisos.", file=sys.stderr)
    for a in areas:
        crit = {c: sum(1 for x in a["criterios"] if x["ciclo"] == c) for c in a["ciclos"]}
        cont = {c: sum(1 for x in a["contenidos"] if x["ciclo"] == c) for c in a["ciclos"]}
        print(f"  {a['codigo']}: {len(a['competencias'])} competencias; criterios {crit}; "
              f"contenidos {cont}", file=sys.stderr)


if __name__ == "__main__":
    main()

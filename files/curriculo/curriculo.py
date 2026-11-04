#!/usr/bin/env python3
"""curriculo: consulta del currículo de Educación Primaria de la Comunidad de Madrid.

Solo usa la biblioteca estándar. Lee el JSON de $CURRICULO_JSON o, si no está
definida, de /usr/share/mint-docente/curriculo/primaria-madrid.json.

Códigos de salida: 0 todo bien; 1 códigos desconocidos o sin resultados;
2 uso incorrecto.
"""
import argparse
import json
import os
import sys
import unicodedata

RUTA_POR_DEFECTO = "/usr/share/mint-docente/curriculo/primaria-madrid.json"
TIPOS = ("competencias", "criterios", "contenidos", "relaciones", "todo")
CICLOS = ("C1", "C2", "C3", "E")


class UsoIncorrecto(Exception):
    """Error de uso: la CLI sale con código 2."""


_cache = {}


def cargar():
    ruta = os.environ.get("CURRICULO_JSON") or RUTA_POR_DEFECTO
    if ruta not in _cache:
        try:
            with open(ruta, encoding="utf-8") as f:
                _cache[ruta] = json.load(f)
        except OSError as e:
            raise SystemExit(f"No se puede leer el currículo ({ruta}): {e.strerror}")
    return _cache[ruta]


def normalizar(texto):
    """Minúsculas y sin tildes ni diéresis (NFKD)."""
    d = unicodedata.normalize("NFKD", texto)
    return "".join(c for c in d if not unicodedata.combining(c)).lower()


def resolver_area(nombre):
    """Áreas cuyo código, nombre o alias coincide (normalizado) con `nombre`."""
    n = normalizar(nombre.strip())
    return [a for a in cargar()["areas"]
            if n in {normalizar(a["codigo"]), normalizar(a["nombre"]),
                     *(normalizar(x) for x in a.get("alias", []))}]


def ciclo_para(area, curso=None, ciclo=None):
    """Ciclo (C1, C2, C3 o E) que corresponde a un curso o ciclo en el área.

    Sin curso ni ciclo devuelve None (todos los ciclos). Lanza ValueError con
    un mensaje en español si el curso o el ciclo no son válidos para el área.
    """
    ciclos = area["ciclos"]
    if curso is not None:
        if not 1 <= curso <= 6:
            raise ValueError(f"El curso debe ser de 1 a 6 (recibido: {curso}).")
        if ciclos == ["E"]:
            return "E"
        c = "C1" if curso <= 2 else "C2" if curso <= 4 else "C3"
        if c not in ciclos:
            if area["codigo"] == "VCE":
                raise ValueError("Valores Cívicos y Éticos solo se imparte en 5º "
                                 f"(recibido: {curso}º).")
            raise ValueError(f"{area['nombre']} no se imparte en {curso}º.")
        if area["codigo"] == "VCE" and curso != 5:
            raise ValueError("Valores Cívicos y Éticos solo se imparte en 5º "
                             f"(recibido: {curso}º).")
        return c
    if ciclo is not None:
        c = ciclo.upper()
        if c not in ciclos:
            raise ValueError(f"{area['nombre']} ({area['codigo']}) no tiene el ciclo "
                             f"{ciclo}; sus ciclos son: {', '.join(ciclos)}.")
        return c
    return None


def indice_codigos():
    """Todos los códigos válidos: competencias, criterios y descriptores."""
    datos = cargar()
    codigos = {d["codigo"] for d in datos["descriptores"]}
    for a in datos["areas"]:
        codigos.update(c["codigo"] for c in a["competencias"])
        codigos.update(c["codigo"] for c in a["criterios"])
    return codigos


# ---------------------------------------------------------------- utilidades

def texto_competencia(comp, ciclo):
    """Redacción de la competencia para un ciclo (la del primer ciclo si no varía)."""
    return comp.get("variantes", {}).get(ciclo, comp["texto"]) if ciclo else comp["texto"]


def una_area(nombre):
    areas = resolver_area(nombre)
    if not areas:
        raise UsoIncorrecto(f"No conozco el área «{nombre}». Usa «curriculo areas» "
                            "para ver las que hay.")
    if len(areas) > 1:
        opciones = "\n".join(f"- {a['codigo']}: {a['nombre']}" for a in areas)
        raise UsoIncorrecto(f"«{nombre}» es ambiguo. Elige una de estas áreas "
                            f"(usa su código):\n{opciones}")
    return areas[0]


def curso_o_ciclo(area, args):
    if args.curso is not None and args.ciclo is not None:
        raise UsoIncorrecto("Usa --curso o --ciclo, no los dos a la vez.")
    try:
        return ciclo_para(area, args.curso, args.ciclo)
    except ValueError as e:
        raise UsoIncorrecto(str(e))


def lista_contenidos(area, ciclo):
    """Líneas Markdown de los contenidos agrupados por ciclo, bloque y sub-bloque."""
    out, ult = [], None
    for c in area["contenidos"]:
        if ciclo and c["ciclo"] != ciclo:
            continue
        clave = (c["ciclo"], c["bloque"], c.get("subbloque"))
        if ult is None or clave[:2] != ult[:2]:
            if ciclo is None or ult is None or clave[0] != ult[0]:
                if ciclo is None:
                    out += ["", f"### Ciclo {c['ciclo']}"]
            out += ["", f"#### {c['bloque']}", ""]
        if clave != ult and c.get("subbloque"):
            out += ["", f"**{c['subbloque']}**", ""]
        ult = clave
        out.append("  " * max(c.get("nivel", 1) - 1, 0) + f"- {c['texto']}")
    return out


def lista_competencias(area, ciclo):
    out = []
    for comp in area["competencias"]:
        out += ["", f"### {comp['codigo']}", "", texto_competencia(comp, ciclo)]
        if ciclo is None:
            for cic, txt in comp.get("variantes", {}).items():
                out += ["", f"- Redacción en {cic}: {txt}"]
        out += ["", f"Descriptores: {', '.join(comp['descriptores'])}"]
        if comp.get("equivalencia_rd"):
            out.append(f"Equivalencia en el RD 157/2022: {comp['equivalencia_rd']}")
    return out


def criterios_de(area, ciclo, comp=None):
    return [c for c in area["criterios"]
            if (ciclo is None or c["ciclo"] == ciclo)
            and (comp is None or c["competencia"] == comp)]


def lista_criterios(area, ciclo):
    out = []
    for comp in area["competencias"]:
        crit = criterios_de(area, ciclo, comp["codigo"])
        if crit:
            out += ["", f"### {comp['codigo']}", ""]
            out += [f"- `{c['codigo']}` — {c['texto']}" for c in crit]
    return out


def lista_relaciones(area, ciclo):
    out = []
    for comp in area["competencias"]:
        out += ["", f"### {comp['codigo']} — {texto_competencia(comp, ciclo)}", "",
                f"- Descriptores: {', '.join(comp['descriptores'])}"]
        if ciclo is None:
            out += [f"- Redacción en {cic}: {txt}"
                    for cic, txt in comp.get("variantes", {}).items()]
        if comp.get("equivalencia_rd"):
            out.append(f"- Equivalencia en el RD 157/2022: {comp['equivalencia_rd']}")
        out.append("- Criterios:")
        crit = criterios_de(area, ciclo, comp["codigo"])
        out += [f"  - `{c['codigo']}` — {c['texto']}" for c in crit] or ["  - (ninguno)"]
    return out


# ---------------------------------------------------------------- subórdenes

def cmd_areas(args):
    print("| Código | Nombre | Ciclos | Alias |\n|---|---|---|---|")
    for a in cargar()["areas"]:
        print(f"| {a['codigo']} | {a['nombre']} | {', '.join(a['ciclos'])} | "
              f"{', '.join(a.get('alias', []))} |")
    return 0


def cmd_ver(args):
    area = una_area(args.area)
    ciclo = curso_o_ciclo(area, args)
    t = args.tipo
    titulo = f"# {area['codigo']} — {area['nombre']}" + (f" (ciclo {ciclo})" if ciclo else "")
    out = [titulo]
    if t in ("competencias", "todo"):
        out += ["", "## Competencias específicas"] + lista_competencias(area, ciclo)
    if t in ("criterios", "todo"):
        out += ["", "## Criterios de evaluación"] + lista_criterios(area, ciclo)
    if t in ("contenidos", "todo"):
        out += ["", "## Contenidos"] + lista_contenidos(area, ciclo)
    if t == "relaciones":
        out += ["", "## Relaciones"] + lista_relaciones(area, ciclo)
    print("\n".join(out))
    return 0


def cmd_buscar(args):
    n = normalizar(args.texto.strip())
    if not n:
        raise UsoIncorrecto("Indica qué buscar.")
    if args.curso is not None and args.ciclo is not None:
        raise UsoIncorrecto("Usa --curso o --ciclo, no los dos a la vez.")
    if args.area:
        areas = [una_area(args.area)]
    else:
        areas = cargar()["areas"]
        if args.curso is not None and not 1 <= args.curso <= 6:
            raise UsoIncorrecto(f"El curso debe ser de 1 a 6 (recibido: {args.curso}).")
    salida = []
    for a in areas:
        try:
            ciclo = ciclo_para(a, args.curso, args.ciclo)
        except ValueError as e:
            if args.area:
                raise UsoIncorrecto(str(e))
            continue  # sin área concreta, se omiten las áreas a las que no aplica
        for comp in a["competencias"]:
            if ciclo:
                txt = texto_competencia(comp, ciclo)
                if n in normalizar(txt):
                    salida.append(f"{comp['codigo']} — {txt}")
            else:
                if n in normalizar(comp["texto"]):
                    salida.append(f"{comp['codigo']} — {comp['texto']}")
                for cic, txt in comp.get("variantes", {}).items():
                    if n in normalizar(txt):
                        salida.append(f"{comp['codigo']} ({cic}) — {txt}")
        for c in criterios_de(a, ciclo):
            if n in normalizar(c["texto"]):
                salida.append(f"{c['codigo']} — {c['texto']}")
        for c in a["contenidos"]:
            if (ciclo is None or c["ciclo"] == ciclo) and n in normalizar(c["texto"]):
                sub = f" › {c['subbloque']}" if c.get("subbloque") else ""
                salida.append(f"{a['codigo']} {c['ciclo']} · {c['bloque']}{sub} — {c['texto']}")
    if not salida:
        print(f"No hay resultados para «{args.texto}».")
        return 1
    print("\n".join(salida))
    return 0


def cmd_descriptores(args):
    descs = cargar()["descriptores"]
    if args.clave:
        clave = args.clave.upper()
        descs = [d for d in descs if d["competencia_clave"].upper() == clave
                 or d["codigo"].upper() == clave]
        if not descs:
            print(f"Competencia clave o descriptor desconocido: «{args.clave}».")
            return 1
    ult = None
    for d in descs:
        if d["competencia_clave"] != ult:
            ult = d["competencia_clave"]
            print(f"\n## {ult}\n")
        print(f"- `{d['codigo']}` — {d['texto']}")
    return 0


def cmd_fuentes(args):
    print("# Fuentes normativas\n")
    for f in cargar()["fuentes"]:
        print(f"- {f['norma']} ({f.get('boletin', '')}, {f.get('fecha', '')})")
        print(f"  - URL: {f['url']}")
        if f.get("sha256"):
            print(f"  - sha256: {f['sha256']}")
    return 0


def parser():
    p = argparse.ArgumentParser(prog="curriculo",
                                description="Consulta el currículo de Primaria de la Comunidad de Madrid.")
    sub = p.add_subparsers(dest="orden", required=True, metavar="orden")
    sub.add_parser("areas", help="lista las áreas, sus ciclos y alias").set_defaults(f=cmd_areas)

    def seleccion(s):
        s.add_argument("--ciclo", choices=CICLOS, type=str.upper, help="C1, C2, C3 o E")
        s.add_argument("--curso", type=int, help="curso de 1 a 6 (se traduce a ciclo)")

    v = sub.add_parser("ver", help="muestra competencias, criterios y contenidos de un área")
    v.add_argument("area", help="código, nombre o alias (mates, lengua, plástica...)")
    seleccion(v)
    v.add_argument("--tipo", choices=TIPOS, default="todo")
    v.set_defaults(f=cmd_ver)
    b = sub.add_parser("buscar", help="busca un texto sin distinguir tildes ni mayúsculas")
    b.add_argument("texto")
    b.add_argument("--area")
    seleccion(b)
    b.set_defaults(f=cmd_buscar)
    d = sub.add_parser("descriptores", help="descriptores del Perfil de salida")
    d.add_argument("clave", nargs="?", help="competencia clave (CCL, STEM...)")
    d.set_defaults(f=cmd_descriptores)
    sub.add_parser("fuentes", help="normas, fechas y URL").set_defaults(f=cmd_fuentes)
    return p


def main(argv=None):
    p = parser()
    args = p.parse_args(argv)
    try:
        return args.f(args)
    except UsoIncorrecto as e:
        print(f"Error: {e}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())

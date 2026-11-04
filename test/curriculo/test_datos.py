"""Pruebas de integridad del currículo de Primaria de Madrid (files/curriculo/).

Se ejecutan en el anfitrión:
    .venv-dev/bin/python -m unittest discover -s test/curriculo -p 'test_datos.py'
"""
import json
import re
import unittest
from pathlib import Path

RAIZ = Path(__file__).resolve().parents[2]
DATOS = RAIZ / "files" / "curriculo" / "primaria-madrid.json"
RECUENTOS = RAIZ / "files" / "curriculo" / "recuentos-esperados.json"

PREFIJOS = "CN|CS|EA|EF|LCL|ING|MAT|VCE|SLE|TYR"
RE_COMPETENCIA = re.compile(rf"^({PREFIJOS})\.CE[0-9]+$")
RE_CRITERIO = re.compile(rf"^({PREFIJOS})\.[0-9]+\.[0-9]+\.(C1|C2|C3|E)$")
RE_DESCRIPTOR = re.compile(r"^(CCL|CP|STEM|CD|CPSAA|CC|CE|CCEC)[0-9]$")

AREAS_Y_CICLOS = {
    "CN": ["C1", "C2", "C3"],
    "CS": ["C1", "C2", "C3"],
    "EA": ["C1", "C2", "C3"],
    "EF": ["C1", "C2", "C3"],
    "LCL": ["C1", "C2", "C3"],
    "ING": ["C1", "C2", "C3"],
    "MAT": ["C1", "C2", "C3"],
    "VCE": ["C3"],
    "SLE": ["E"],
    "TYR": ["E"],
}

RESTOS_PROHIBIDOS = [
    "BOCM-20220718-1",
    "BOLETÍN OFICIAL DE LA COMUNIDAD DE MADRID",
    "B.O.C.M. Núm.",
    "Pág. ",
    "DE JULIO DE 2022",
]


class TestDatosCurriculo(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        with open(DATOS, encoding="utf-8") as f:
            cls.datos = json.load(f)
        cls.areas = {a["codigo"]: a for a in cls.datos["areas"]}

    def _textos(self):
        """Todos los textos del archivo, con una etiqueta para localizarlos."""
        for d in self.datos["descriptores"]:
            yield d["codigo"], d["texto"]
        for a in self.datos["areas"]:
            for c in a["competencias"]:
                yield c["codigo"], c["texto"]
                for ciclo, texto in c["variantes"].items():
                    yield f"{c['codigo']} ({ciclo})", texto
            for c in a["criterios"]:
                yield c["codigo"], c["texto"]
            for i, c in enumerate(a["contenidos"]):
                etiqueta = f"{a['codigo']} contenido {i} (pág. {c['pagina']})"
                yield etiqueta, c["texto"]
                yield etiqueta + " bloque", c["bloque"]
                if c["subbloque"] is not None:
                    yield etiqueta + " subbloque", c["subbloque"]

    def test_codigos_unicos(self):
        codigos = [d["codigo"] for d in self.datos["descriptores"]]
        codigos += [a["codigo"] for a in self.datos["areas"]]
        for a in self.datos["areas"]:
            codigos += [c["codigo"] for c in a["competencias"]]
            codigos += [c["codigo"] for c in a["criterios"]]
        repetidos = sorted({c for c in codigos if codigos.count(c) > 1})
        self.assertEqual(repetidos, [], f"Códigos repetidos: {repetidos}")

    def test_formato_codigos(self):
        for d in self.datos["descriptores"]:
            self.assertRegex(d["codigo"], RE_DESCRIPTOR)
        for a in self.datos["areas"]:
            for c in a["competencias"]:
                self.assertRegex(c["codigo"], RE_COMPETENCIA)
                self.assertTrue(c["codigo"].startswith(a["codigo"] + "."), c["codigo"])
            for c in a["criterios"]:
                self.assertRegex(c["codigo"], RE_CRITERIO)
                self.assertTrue(c["codigo"].startswith(a["codigo"] + "."), c["codigo"])
                self.assertTrue(c["codigo"].endswith("." + c["ciclo"]), c["codigo"])

    def test_areas_y_ciclos(self):
        self.assertEqual(
            {a: self.areas[a]["ciclos"] for a in self.areas}, AREAS_Y_CICLOS
        )
        self.assertEqual(len(self.datos["areas"]), len(AREAS_Y_CICLOS))
        for codigo, a in self.areas.items():
            with self.subTest(area=codigo):
                self.assertTrue(a["nombre"])
                self.assertTrue(a["alias"])
                self.assertTrue(a["competencias"])
                for c in a["competencias"]:
                    # Redacciones de la competencia que cambian en algún ciclo.
                    for ciclo, texto in c["variantes"].items():
                        self.assertIn(ciclo, a["ciclos"], c["codigo"])
                        self.assertNotEqual(texto, c["texto"], c["codigo"])
                for c in a["criterios"]:
                    self.assertIn(c["ciclo"], a["ciclos"], c["codigo"])
                    self.assertIsInstance(c["pagina"], int)
                for c in a["contenidos"]:
                    self.assertIn(c["ciclo"], a["ciclos"])
                    self.assertIn(c["nivel"], (1, 2, 3))
                    self.assertIsInstance(c["pagina"], int)
                for ciclo in a["ciclos"]:
                    self.assertTrue(
                        any(c["ciclo"] == ciclo for c in a["criterios"]),
                        f"{codigo} no tiene criterios en {ciclo}",
                    )
                    self.assertTrue(
                        any(c["ciclo"] == ciclo for c in a["contenidos"]),
                        f"{codigo} no tiene contenidos en {ciclo}",
                    )

    def test_criterio_apunta_a_competencia(self):
        for codigo, a in self.areas.items():
            competencias = {c["codigo"]: c for c in a["competencias"]}
            for cr in a["criterios"]:
                with self.subTest(criterio=cr["codigo"]):
                    self.assertIn(cr["competencia"], competencias)
                    n = int(cr["codigo"].split(".")[1])
                    self.assertEqual(n, competencias[cr["competencia"]]["numero"])
            for c in a["competencias"]:
                self.assertEqual(c["codigo"], f"{codigo}.CE{c['numero']}")

    def test_descriptores_existen(self):
        conocidos = {d["codigo"] for d in self.datos["descriptores"]}
        for d in self.datos["descriptores"]:
            self.assertTrue(d["codigo"].startswith(d["competencia_clave"]), d["codigo"])
            self.assertRegex(d["codigo"], rf"^{d['competencia_clave']}[0-9]$")
        for a in self.datos["areas"]:
            for c in a["competencias"]:
                for d in c["descriptores"]:
                    self.assertIn(d, conocidos, f"{c['codigo']} cita {d}")

    def test_equivalencias_cs_cn(self):
        cs = {c["numero"]: c["equivalencia_rd"] for c in self.areas["CS"]["competencias"]}
        self.assertEqual(cs, {n: f"CMNSC.CE{n + 4}" for n in range(1, 6)})
        cn = {c["numero"]: c["equivalencia_rd"] for c in self.areas["CN"]["competencias"]}
        self.assertEqual(cn, {n: f"CMNSC.CE{n}" for n in range(1, 7)})
        for codigo, a in self.areas.items():
            if codigo in ("CN", "CS"):
                continue
            for c in a["competencias"]:
                eq = c["equivalencia_rd"]
                if eq is not None:
                    self.assertRegex(eq, rf"\.CE{c['numero']}$", c["codigo"])

    def test_recuentos(self):
        with open(RECUENTOS, encoding="utf-8") as f:
            esperados = json.load(f)
        if "_estado" in esperados:
            self.skipTest(esperados["_estado"])
        for codigo, esp in esperados.items():
            a = self.areas[codigo]
            with self.subTest(area=codigo):
                self.assertEqual(len(a["competencias"]), esp["competencias"])
                criterios = {
                    ciclo: sum(1 for c in a["criterios"] if c["ciclo"] == ciclo)
                    for ciclo in a["ciclos"]
                }
                self.assertEqual(criterios, esp["criterios"])
                contenidos = {
                    ciclo: sum(1 for c in a["contenidos"] if c["ciclo"] == ciclo)
                    for ciclo in a["ciclos"]
                }
                self.assertEqual(contenidos, esp["contenidos"])

    def test_textos_limpios(self):
        for etiqueta, texto in self._textos():
            with self.subTest(donde=etiqueta):
                self.assertIsInstance(texto, str)
                self.assertTrue(texto.strip(), "texto vacío")
                self.assertNotRegex(texto, r"\w-$", "termina en guion de corte")
                for resto in RESTOS_PROHIBIDOS:
                    self.assertNotIn(resto, texto)
                self.assertFalse(texto.startswith("x "), "viñeta «x» sin quitar")
                self.assertFalse(texto.startswith(""), "viñeta «» sin quitar")

    def test_fuentes(self):
        fuentes = self.datos["fuentes"]
        self.assertEqual(len(fuentes), 4)
        for f in fuentes:
            with self.subTest(norma=f["norma"]):
                for campo in ("norma", "boletin", "fecha"):
                    self.assertTrue(f[campo])
                self.assertTrue(f["url"].startswith("https://"), f["url"])
                self.assertRegex(f["sha256"], r"^[0-9a-f]{64}$")
        self.assertRegex(self.datos["version"], r"^\d{4}-\d{2}-\d{2}$")


if __name__ == "__main__":
    unittest.main()

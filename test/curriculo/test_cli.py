"""Pruebas de la CLI `curriculo` (files/curriculo/curriculo.py).

Se ejecutan en el anfitrión:
    .venv-dev/bin/python -m unittest test.curriculo.test_cli
"""
import json
import os
import subprocess
import sys
import unittest
from pathlib import Path

RAIZ = Path(__file__).resolve().parents[2]
CLI = RAIZ / "files" / "curriculo" / "curriculo.py"
DATOS = RAIZ / "files" / "curriculo" / "primaria-madrid.json"
CODIGOS_AREA = ["CN", "CS", "EA", "EF", "LCL", "ING", "MAT", "VCE", "SLE", "TYR"]

with open(DATOS, encoding="utf-8") as f:
    JSON = json.load(f)
AREAS = {a["codigo"]: a for a in JSON["areas"]}


def curriculo(*args):
    env = dict(os.environ, CURRICULO_JSON=str(DATOS))
    return subprocess.run([sys.executable, str(CLI), *args], capture_output=True,
                          text=True, encoding="utf-8", env=env)


def lineas(r):
    return [l for l in r.stdout.splitlines() if l.strip()]


class TestCli(unittest.TestCase):
    def test_areas_lista_todas(self):
        r = curriculo("areas")
        self.assertEqual(r.returncode, 0, r.stderr)
        for c in CODIGOS_AREA:
            self.assertIn(c, r.stdout)

    def test_ver_criterios_ciclo(self):
        r = curriculo("ver", "MAT", "--ciclo", "C2", "--tipo", "criterios")
        self.assertEqual(r.returncode, 0, r.stderr)
        for c in AREAS["MAT"]["criterios"]:
            if c["ciclo"] == "C2":
                self.assertIn(c["codigo"], r.stdout)
        self.assertNotIn(".C1", r.stdout)
        self.assertNotIn(".C3", r.stdout)

    def test_curso_a_ciclo(self):
        a = curriculo("ver", "MAT", "--curso", "3")
        b = curriculo("ver", "MAT", "--ciclo", "C2")
        self.assertEqual(a.returncode, 0, a.stderr)
        self.assertEqual(a.stdout, b.stdout)
        for n in ("7", "0"):
            r = curriculo("ver", "MAT", "--curso", n)
            self.assertEqual(r.returncode, 2)
            self.assertIn("1 a 6", r.stderr + r.stdout)
        r = curriculo("ver", "TYR", "--curso", "2", "--tipo", "criterios")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn(".E", r.stdout)
        r = curriculo("ver", "VCE", "--curso", "3")
        self.assertEqual(r.returncode, 2)
        self.assertIn("5º", r.stderr + r.stdout)

    def test_alias(self):
        esperado = {
            "mates": "MAT", "Matemáticas": "MAT", "MAT": "MAT",
            "ingles": "ING", "Inglés": "ING", "lengua": "LCL",
            "sociales": "CS", "naturales": "CN", "educación física": "EF",
            "plástica": "EA", "música": "EA", "robótica": "TYR",
            "francés": "SLE", "valores": "VCE",
        }
        sys.path.insert(0, str(CLI.parent))
        try:
            import curriculo as modulo
        finally:
            sys.path.pop(0)
        os.environ["CURRICULO_JSON"] = str(DATOS)
        for nombre, codigo in esperado.items():
            with self.subTest(nombre=nombre):
                self.assertEqual([a["codigo"] for a in modulo.resolver_area(nombre)], [codigo])
        self.assertEqual(modulo.resolver_area("zzz"), [])
        self.assertEqual(modulo.normalizar("DivisiÓn Ünico"), "division unico")
        # Y por la CLI, con la misma salida
        self.assertEqual(curriculo("ver", "mates", "--ciclo", "C1").stdout,
                         curriculo("ver", "MAT", "--ciclo", "C1").stdout)

    def test_alias_ambiguo(self):
        r = curriculo("ver", "ciencias")
        self.assertEqual(r.returncode, 2)
        self.assertTrue("CN" in r.stderr + r.stdout and "CS" in r.stderr + r.stdout)
        r = curriculo("ver", "idioma")
        self.assertEqual(r.returncode, 2)
        self.assertTrue("ING" in r.stderr + r.stdout and "SLE" in r.stderr + r.stdout)

    def test_buscar_sin_tildes(self):
        a = curriculo("buscar", "division", "--area", "MAT")
        b = curriculo("buscar", "DIVISIÓN", "--area", "MAT")
        self.assertEqual(a.returncode, 0, a.stderr)
        self.assertTrue(lineas(a))
        self.assertEqual(lineas(a), lineas(b))
        r = curriculo("buscar", "lectura", "--area", "lengua")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(lineas(r))

    def test_buscar_contenidos(self):
        r = curriculo("buscar", "vertebrados", "--area", "naturales", "--curso", "1")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue([l for l in lineas(r) if "A. Cultura científica" in l and "CN C1" in l], r.stdout)

    def test_buscar_sin_resultados(self):
        r = curriculo("buscar", "xyzzyqwerty")
        self.assertEqual(r.returncode, 1)
        self.assertIn("No hay resultados", r.stdout + r.stderr)

    def test_relaciones(self):
        r = curriculo("ver", "CS", "--ciclo", "C1", "--tipo", "relaciones")
        self.assertEqual(r.returncode, 0, r.stderr)
        for ce in AREAS["CS"]["competencias"]:
            self.assertIn(ce["codigo"], r.stdout)
            self.assertIn(ce["equivalencia_rd"], r.stdout)
            bloque = r.stdout.split(f"### {ce['codigo']}")[1].split("### ")[0]
            self.assertIn(".C1", bloque)

    def test_variantes_por_ciclo(self):
        c3 = curriculo("ver", "LCL", "--ciclo", "C3", "--tipo", "competencias")
        c1 = curriculo("ver", "LCL", "--ciclo", "C1", "--tipo", "competencias")
        todo = curriculo("ver", "LCL", "--tipo", "competencias")
        self.assertIn("de manera creativa", c3.stdout)
        self.assertNotIn("de manera creativa", c1.stdout)
        self.assertIn("de manera creativa", todo.stdout)  # listada como variante
        self.assertIn("C3", todo.stdout)
        # buscar en un ciclo concreto muestra la redacción de ese ciclo
        b = curriculo("buscar", "de manera creativa", "--area", "LCL", "--ciclo", "C3")
        self.assertTrue([l for l in lineas(b) if l.startswith("LCL.CE6")], b.stdout)

    def test_ver_contenidos_agrupados(self):
        r = curriculo("ver", "EA", "--ciclo", "C1", "--tipo", "contenidos")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("BLOQUE I Música y danza", r.stdout)
        self.assertIn("A. Recepción y análisis", r.stdout)

    def test_descriptores(self):
        r = curriculo("descriptores", "STEM")
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertIn("STEM2", r.stdout)
        self.assertNotIn("CCL1", r.stdout)
        self.assertEqual(curriculo("descriptores", "XXX").returncode, 1)

    def test_fuentes(self):
        r = curriculo("fuentes")
        self.assertEqual(r.returncode, 0, r.stderr)
        for f in JSON["fuentes"]:
            self.assertIn(f["url"], r.stdout)


if __name__ == "__main__":
    unittest.main()

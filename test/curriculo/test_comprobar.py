"""Pruebas de `curriculo comprobar` (files/curriculo/curriculo.py).

    .venv-dev/bin/python -m unittest test.curriculo.test_comprobar
Los archivos de prueba se generan en un directorio temporal.
"""
import json
import tempfile
import unittest
from pathlib import Path

from test.curriculo.test_cli import DATOS, curriculo

with open(DATOS, encoding="utf-8") as f:
    JSON = json.load(f)
MAT = next(a for a in JSON["areas"] if a["codigo"] == "MAT")
CRIT_MAT_3_1 = sorted(c["codigo"] for c in MAT["criterios"]
                      if c["codigo"].startswith("MAT.3.1."))


class ComprobarBase(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.dir = Path(self._tmp.name)
        self.addCleanup(self._tmp.cleanup)

    def md(self, nombre, texto):
        p = self.dir / nombre
        p.write_text(texto, encoding="utf-8")
        return str(p)


class TestComprobar(ComprobarBase):
    def test_md_valido(self):
        p = self.md("a.md", "Trabajamos MAT.3.1.C2, CS.CE1 y CCL1 en la unidad.")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("Todos los códigos (3) existen en el currículo de Madrid", r.stdout)

    def test_codigo_inventado(self):
        p = self.md("a.md", "Criterio MAT.9.9.C2 y MAT.3.1.C2.")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 1)
        self.assertIn("MAT.9.9.C2", r.stdout)
        self.assertNotIn("MAT.3.1.C2", r.stdout)

    def test_codigo_estatal_sin_ciclo(self):
        p = self.md("a.md", "Evaluamos con MAT.3.1 esta semana.")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 1)
        self.assertGreater(len(CRIT_MAT_3_1), 0)
        for c in CRIT_MAT_3_1:
            self.assertIn(c, r.stdout)
        sugeridos = {w.strip(",.;?") for w in r.stdout.split() if w.startswith("MAT.3.1.")}
        self.assertEqual(sugeridos, set(CRIT_MAT_3_1))

    def test_numeracion_rd(self):
        p = self.md("a.md", "Competencia CMNSC.CE7 del RD.")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 1)
        self.assertIn("CMNSC.CE7", r.stdout)
        self.assertIn("CS.CE3", r.stdout)

    def test_ce_suelto_y_forma_con_puntos_son_sospechosos(self):
        p = self.md("a.md", "Ver CE7 y también C.E. 3.")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 1)
        self.assertIn("CE7", r.stdout)
        self.assertIn("C.E. 3", r.stdout)

    def test_docx_en_tabla(self):
        import docx
        d = docx.Document()
        d.add_paragraph("Programación")
        t = d.add_table(rows=2, cols=2)
        t.cell(0, 0).text = "MAT.3.1.C2"
        t.cell(1, 1).text = "LCL.7.7.C7"
        ruta = self.dir / "prog.docx"
        d.save(str(ruta))
        r = curriculo("comprobar", str(ruta))
        self.assertEqual(r.returncode, 1, r.stdout + r.stderr)
        self.assertIn("LCL.7.7.C7", r.stdout)
        self.assertNotIn("MAT.3.1.C2", r.stdout)

    def test_xlsx(self):
        import openpyxl
        wb = openpyxl.Workbook()
        ws = wb.active
        ws["A1"] = "MAT.3.1.C2"
        ws["B2"] = "CS.CE2"
        ws["C3"] = "MAT.9.9.C1"
        ruta = self.dir / "prog.xlsx"
        wb.save(str(ruta))
        r = curriculo("comprobar", str(ruta))
        self.assertEqual(r.returncode, 1, r.stdout + r.stderr)
        self.assertIn("MAT.9.9.C1", r.stdout)
        self.assertNotIn("CS.CE2", r.stdout)

    def test_varios_archivos(self):
        a = self.md("a.md", "MAT.3.1.C2 y MAT.9.9.C2")
        b = self.md("b.md", "CS.CE1 y LCL.8.8.C1")
        r = curriculo("comprobar", a, b)
        self.assertEqual(r.returncode, 1)
        lin = r.stdout.splitlines()
        la = [l for l in lin if "MAT.9.9.C2" in l][0]
        lb = [l for l in lin if "LCL.8.8.C1" in l][0]
        self.assertIn("a.md", la)
        self.assertNotIn("b.md", la)
        self.assertIn("b.md", lb)
        self.assertNotIn("a.md", lb)

    def test_archivo_inexistente(self):
        r = curriculo("comprobar", str(self.dir / "no_existe.md"))
        self.assertEqual(r.returncode, 2)
        self.assertIn("no_existe.md", r.stderr + r.stdout)

    def test_siglas_ajenas_no_son_codigos(self):
        p = self.md("a.md", "UD.3.1, PDF.1.2, ESO.1.2, LOE, Art. 6.6, RD 157/2022")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 0, r.stdout)
        self.assertIn("No se ha citado ningún código del currículo", r.stdout)

    def test_docx_negrita_parcial(self):
        import docx
        d = docx.Document()
        par = d.add_paragraph()
        par.add_run("MAT.3.1.").bold = True
        par.add_run("C2")
        ruta = self.dir / "neg.docx"
        d.save(str(ruta))
        r = curriculo("comprobar", str(ruta))
        self.assertEqual(r.returncode, 0, r.stdout + r.stderr)
        self.assertIn("(1)", r.stdout)

    def test_docx_corrupto(self):
        ruta = self.dir / "roto.docx"
        ruta.write_bytes(b"\x00\x01basura que no es un zip")
        r = curriculo("comprobar", str(ruta))
        self.assertEqual(r.returncode, 2, r.stdout)
        self.assertIn("roto.docx", r.stdout + r.stderr)

    def test_sin_codigos(self):
        p = self.md("a.md", "Hoy hablamos de fracciones y del río Tajo (MP3).")
        r = curriculo("comprobar", p)
        self.assertEqual(r.returncode, 0)
        self.assertIn("No se ha citado ningún código del currículo", r.stdout)


class TestExtraer(unittest.TestCase):
    def test_extraer(self):
        import sys
        from test.curriculo.test_cli import CLI
        sys.path.insert(0, str(CLI.parent))
        import curriculo as c
        self.assertEqual(c.extraer_codigos("MAT.3.1.C2, MAT.3.1.C2 y TYR.2.4.E."),
                         ["MAT.3.1.C2", "TYR.2.4.E"])


if __name__ == "__main__":
    unittest.main()

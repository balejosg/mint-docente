"""Pruebas de la orden `compartir`, sin red.

Dos «docentes» (ana y luis) con su propia carpeta de espacios sobre un
«GitHub» falso: repositorios bare locales + test/falso-gh.

    python3 -m unittest test/test_compartir.py -v
"""
import json
from typing import Any
import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

AQUI = Path(__file__).resolve().parent
COMPARTIR = Path(os.environ.get("COMPARTIR_BIN", AQUI.parent / "files" / "compartir" / "compartir"))
FALSO_GH = AQUI / "falso-gh"
CLAVE_FALSA = "sk-ant-" + "a1B2c3D4e5" * 4


class Base(unittest.TestCase):
    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp(prefix="compartir-"))
        self.base = self.tmp / "github"
        self.base.mkdir()
        self.log = self.tmp / "gh.log"

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def raiz(self, usuario):
        return self.tmp / "aula" / usuario / "Compartido"

    def skills(self, usuario):
        return self.tmp / "skills" / usuario

    def entorno(self, usuario):
        e = {k: v for k, v in os.environ.items() if not k.startswith(("GIT_", "GH_"))}
        e.update({
            "HOME": str(self.tmp / "home" / usuario),
            "GIT_CONFIG_GLOBAL": "/dev/null",
            "GIT_CONFIG_NOSYSTEM": "1",
            "COMPARTIR_GH": str(FALSO_GH),
            "COMPARTIR_GIT_BASE": str(self.base),
            "COMPARTIR_RAIZ": str(self.raiz(usuario)),
            "COMPARTIR_SKILLS": str(self.skills(usuario)),
            "FALSO_GH_USUARIO": usuario,
            "FALSO_GH_DIR": str(self.tmp / "estado-gh"),
            "FALSO_GH_LOG": str(self.log),
        })
        (self.tmp / "home" / usuario).mkdir(parents=True, exist_ok=True)
        return e

    def compartir(self, usuario, *args) -> tuple[int, Any, str]:
        r = subprocess.run([str(COMPARTIR), *args], env=self.entorno(usuario),
                           capture_output=True, text=True)
        datos = None
        if r.stdout.strip():
            try:
                datos = json.loads(r.stdout)
            except json.JSONDecodeError:
                datos = r.stdout
        return r.returncode, datos, r.stderr

    def ok(self, usuario, *args) -> Any:
        rc, datos, err = self.compartir(usuario, *args)
        self.assertEqual(rc, 0, f"compartir {args} → {rc}: {err}")
        return datos

    def llamadas(self, usuario=None):
        if not self.log.exists():
            return []
        filas = [json.loads(l) for l in self.log.read_text().splitlines()]
        return [f["args"] for f in filas if usuario is None or f["usuario"] == usuario]

    def git(self, ruta, *args):
        return subprocess.run(["git", "-C", str(ruta), *args], check=True,
                              capture_output=True, text=True,
                              env={**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null"}).stdout

    def remoto_log(self, propietario, slug):
        return self.git(self.base / propietario / f"{slug}.git", "log", "--oneline", "--all")

    def espacio_compartido(self, nombre="Fichas"):
        """ana crea el espacio y luis tiene una copia (sin pasar por invitaciones)."""
        self.ok("ana", "crear", nombre)
        slug = nombre.lower()
        destino = self.raiz("luis") / nombre
        destino.parent.mkdir(parents=True, exist_ok=True)
        subprocess.run(["git", "clone", "-q", str(self.base / "ana" / f"{slug}.git"), str(destino)],
                       check=True, env={**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null"})
        self.git(destino, "config", "user.name", "Luis")
        self.git(destino, "config", "user.email", "1001+luis@users.noreply.github.com")
        self.git(destino, "config", "compartir.usuario", "luis")
        return self.raiz("ana") / nombre, destino


class TestSesionYCrear(Base):
    def test_sin_sesion_codigo_3(self):
        (self.tmp / "estado-gh").mkdir()
        (self.tmp / "estado-gh" / "sin_sesion").touch()
        rc, _, err = self.compartir("ana", "estado")
        self.assertEqual(rc, 3)
        self.assertIn("Conectar GitHub para compartir", err)
        self.assertNotIn("Traceback", err)

    def test_crear_slug_y_carpeta(self):
        datos = self.ok("ana", "crear", "3º ESO Biología")
        carpeta = self.raiz("ana") / "3º ESO Biología"
        self.assertTrue(carpeta.is_dir())
        self.assertEqual(datos["espacio"], "3º ESO Biología")
        self.assertEqual(datos["propietario"], "ana")
        creates = [a for a in self.llamadas() if a[:2] == ["repo", "create"]]
        self.assertEqual(len(creates), 1)
        self.assertEqual(creates[0][2], "3o-eso-biologia")
        self.assertIn("--private", creates[0])
        self.assertIn(["repo", "edit", "ana/3o-eso-biologia", "--add-topic", "mint-docente"],
                      self.llamadas())
        self.assertIn(".~lock.*#", (carpeta / ".gitignore").read_text())
        self.assertTrue(self.git(carpeta, "config", "user.email").strip()
                        .endswith("@users.noreply.github.com"))
        self.assertIn("Espacio", self.remoto_log("ana", "3o-eso-biologia"))

    def test_crear_nombre_ya_existe(self):
        self.ok("ana", "crear", "Fichas")
        rc, _, err = self.compartir("ana", "crear", "Fichas")
        self.assertEqual(rc, 1)
        self.assertIn("Ya tienes un espacio llamado", err)
        creates = [a for a in self.llamadas() if a[:2] == ["repo", "create"]]
        self.assertEqual(len(creates), 1)


class TestSincronizar(Base):
    def test_guardar_y_traer(self):
        ana, luis = self.espacio_compartido()
        (ana / "ficha.docx").write_bytes(b"PK\x03\x04 version ana \x00\x01")
        datos = self.ok("ana", "guardar", "Fichas", "Primera ficha")
        self.assertIn("ficha.docx", datos["subidos"])
        datos = self.ok("luis", "traer", "Fichas")
        esp = datos["espacios"][0]
        self.assertEqual(esp["nombre"], "Fichas")
        self.assertIn("ficha.docx", esp["nuevos"])
        self.assertEqual((luis / "ficha.docx").read_bytes(), (ana / "ficha.docx").read_bytes())

    def test_guardar_texto_se_mezcla(self):
        ana, luis = self.espacio_compartido()
        (ana / "app").mkdir()
        lineas = [f"<p>linea {i}</p>" for i in range(10)]
        (ana / "app" / "index.html").write_text("\n".join(lineas) + "\n")
        self.ok("ana", "guardar", "Fichas", "app")
        self.ok("luis", "traer", "Fichas")
        a = list(lineas); a[0] = "<p>cambio de ana</p>"
        (ana / "app" / "index.html").write_text("\n".join(a) + "\n")
        l = list(lineas); l[9] = "<p>cambio de luis</p>"
        (luis / "app" / "index.html").write_text("\n".join(l) + "\n")
        self.ok("luis", "guardar", "Fichas", "pie")
        datos = self.ok("ana", "guardar", "Fichas", "cabecera")
        self.assertIn("app/index.html", datos["mezclados"])
        self.assertEqual(datos["dos_versiones"], [])
        self.ok("luis", "traer", "Fichas")
        for lado in (ana, luis):
            texto = (lado / "app" / "index.html").read_text()
            self.assertIn("cambio de ana", texto)
            self.assertIn("cambio de luis", texto)
            self.assertNotIn("<<<<<<<", texto)
            self.assertEqual(sorted(p.name for p in (lado / "app").iterdir()), ["index.html"])

    def test_guardar_mismo_archivo_dos_lados(self):
        ana, luis = self.espacio_compartido()
        (ana / "ficha.docx").write_bytes(b"PK original \x00")
        self.ok("ana", "guardar", "Fichas", "original")
        self.ok("luis", "traer", "Fichas")
        (ana / "ficha.docx").write_bytes(b"PK de ana \x00")
        (luis / "ficha.docx").write_bytes(b"PK de luis \x00")
        self.ok("luis", "guardar", "Fichas", "luis")
        datos = self.ok("ana", "guardar", "Fichas", "ana")
        self.assertEqual(datos["dos_versiones"], [["ficha.docx", "ficha (versión de ana).docx"]])
        self.ok("luis", "traer", "Fichas")
        for lado in (ana, luis):
            self.assertEqual((lado / "ficha.docx").read_bytes(), b"PK de luis \x00")
            self.assertEqual((lado / "ficha (versión de ana).docx").read_bytes(), b"PK de ana \x00")

    def test_borrado_frente_a_modificacion(self):
        ana, luis = self.espacio_compartido()
        (ana / "a.docx").write_bytes(b"PK a \x00")
        self.ok("ana", "guardar", "Fichas", "a")
        self.ok("luis", "traer", "Fichas")
        (ana / "a.docx").unlink()
        (luis / "a.docx").write_bytes(b"PK a de luis \x00")
        self.ok("ana", "guardar", "Fichas", "borro a")
        self.ok("luis", "guardar", "Fichas", "cambio a")
        self.ok("ana", "traer", "Fichas")
        for lado in (ana, luis):
            self.assertEqual((lado / "a.docx").read_bytes(), b"PK a de luis \x00")

    def test_guardar_rechaza_claves(self):
        ana, _ = self.espacio_compartido()
        antes = self.remoto_log("ana", "fichas")
        (ana / "app").mkdir()
        (ana / "app" / "script.js").write_text(
            "// mi app\nconst x = 1;\nconst API_KEY = \"" + CLAVE_FALSA + "\";\n")
        rc, _, err = self.compartir("ana", "guardar", "Fichas", "app")
        self.assertEqual(rc, 1)
        self.assertIn("app/script.js:3", err)
        self.assertNotIn(CLAVE_FALSA, err)
        self.assertEqual(self.remoto_log("ana", "fichas"), antes)
        # En .env no molesta: no se sube.
        (ana / "app" / "script.js").write_text("const API_KEY = leerDeEnv();\n")
        (ana / ".env").write_text("API_KEY=\"" + CLAVE_FALSA + "\"\n")
        datos = self.ok("ana", "guardar", "Fichas", "app")
        self.assertNotIn(".env", datos["subidos"])
        self.assertNotIn(CLAVE_FALSA, self.git(self.base / "ana" / "fichas.git", "log", "-p", "--all"))

    def test_guardar_ignora_bloqueos(self):
        ana, _ = self.espacio_compartido()
        (ana / "ficha.docx").write_bytes(b"PK \x00")
        (ana / ".~lock.ficha.docx#").write_text("ana,bloqueo")
        datos = self.ok("ana", "guardar", "Fichas", "ficha")
        self.assertIn("ficha.docx", datos["subidos"])
        self.assertNotIn(".~lock.ficha.docx#", datos["subidos"])

    def test_guardar_archivo_grande(self):
        ana, _ = self.espacio_compartido()
        antes = self.remoto_log("ana", "fichas")
        with open(ana / "video.mp4", "wb") as f:
            f.truncate(96 * 1024 * 1024)
        rc, _, err = self.compartir("ana", "guardar", "Fichas", "video")
        self.assertEqual(rc, 1)
        self.assertIn("video.mp4", err)
        self.assertEqual(self.remoto_log("ana", "fichas"), antes)

    def test_guardar_sin_conexion_y_reintento(self):
        ana, _ = self.espacio_compartido()
        (ana / "ficha.docx").write_bytes(b"PK offline \x00")
        remoto = self.base / "ana" / "fichas.git"
        apartado = self.tmp / "apartado.git"
        remoto.rename(apartado)          # GitHub inalcanzable
        rc, _, err = self.compartir("ana", "guardar", "Fichas", "sin red")
        self.assertEqual(rc, 4)
        self.assertIn("se subirá la próxima vez", err)
        self.assertNotIn("Traceback", err)
        apartado.rename(remoto)
        datos = self.ok("ana", "guardar", "Fichas", "con red")
        self.assertIn("ficha.docx", datos["subidos"])
        self.assertIn("ficha.docx", self.git(remoto, "ls-tree", "-r", "--name-only", "main"))

    def test_estado_pendientes(self):
        ana, _ = self.espacio_compartido()
        (ana / "nuevo.md").write_text("hola\n")
        datos = self.ok("ana", "estado")
        self.assertEqual(datos["usuario"], "ana")
        esp = {e["nombre"]: e for e in datos["espacios"]}
        self.assertEqual(esp["Fichas"]["pendientes"], 1)
        self.assertEqual(esp["Fichas"]["propietario"], "ana")


if __name__ == "__main__":
    unittest.main()

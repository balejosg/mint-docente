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
            "COMPARTIR_SKILLS_SISTEMA": str(AQUI.parent / "files" / "opencode" / "skills"),
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
        self.assertIn("Espacio compartido de Mint Docente: 3º ESO Biología", creates[0])
        self.assertIn(["repo", "edit", "ana/3o-eso-biologia", "--add-topic", "mint-docente"],
                      self.llamadas())
        self.assertIn(".~lock.*#", (carpeta / ".gitignore").read_text())
        self.assertTrue(self.git(carpeta, "config", "user.email").strip()
                        .endswith("@users.noreply.github.com"))
        self.assertIn("Espacio", self.remoto_log("ana", "3o-eso-biologia"))

    def test_crear_nombre_con_barras(self):
        for nombre in ("A/B", "../fuera", ".oculto"):
            rc, _, err = self.compartir("ana", "crear", nombre)
            self.assertEqual(rc, 1, nombre)
            self.assertIn("no puede", err)
        self.assertFalse([a for a in self.llamadas() if a[:2] == ["repo", "create"]])

    def test_espacio_fuera_de_compartido(self):
        self.ok("ana", "crear", "Fichas")
        rc, _, err = self.compartir("ana", "guardar", "..", "x")
        self.assertEqual(rc, 1)
        self.assertIn("No hay ningún espacio", err)

    def test_mensaje_de_error_sin_corchetes(self):
        rc, _, err = self.compartir("ana", "guardar", "NoExiste", "x")
        self.assertEqual(rc, 1)
        self.assertNotIn("['", err)

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
        url = self.git(ana, "remote", "get-url", "origin").strip()
        self.git(ana, "remote", "set-url", "origin", "https://127.0.0.1:9/ana/fichas.git")
        rc, _, err = self.compartir("ana", "guardar", "Fichas", "sin red")
        self.assertEqual(rc, 4)
        self.assertIn("se subirá la próxima vez", err)
        self.assertNotIn("Traceback", err)
        self.git(ana, "remote", "set-url", "origin", url)
        datos = self.ok("ana", "guardar", "Fichas", "con red")
        self.assertIn("ficha.docx", datos["subidos"])
        self.assertIn("ficha.docx", self.git(remoto, "ls-tree", "-r", "--name-only", "main"))

    def test_guardar_sin_acceso_no_es_sin_conexion(self):
        ana, luis = self.espacio_compartido()
        (luis / "nota.md").write_text("hola\n")
        shutil.rmtree(self.base / "ana" / "fichas.git")      # borrado o sin acceso
        rc, _, err = self.compartir("luis", "guardar", "Fichas", "nota")
        self.assertEqual(rc, 1)
        self.assertIn("Ya no tienes acceso a este espacio", err)

    def test_guardar_tras_fetch_sin_mezclar(self):
        ana, luis = self.espacio_compartido()
        (ana / "x.txt").write_text("v1\n")
        self.ok("ana", "guardar", "Fichas", "v1")
        self.ok("luis", "traer", "Fichas")
        (ana / "x.txt").write_text("v2\n")
        self.ok("ana", "guardar", "Fichas", "v2")
        self.git(luis, "fetch", "-q")                       # el agente puede hacer fetch
        (luis / "y.txt").write_text("de luis\n")
        datos = self.ok("luis", "guardar", "Fichas", "y")
        self.assertEqual(datos["subidos"], ["y.txt"])
        self.ok("ana", "traer", "Fichas")
        self.assertEqual((ana / "x.txt").read_text(), "v2\n")
        self.assertEqual((luis / "x.txt").read_text(), "v2\n")

    def test_renombrado_en_los_dos_lados(self):
        ana, luis = self.espacio_compartido()
        (ana / "x.txt").write_text("contenido\n")
        (ana / "otro.md").write_text("a\n")
        self.ok("ana", "guardar", "Fichas", "x")
        self.ok("luis", "traer", "Fichas")
        self.git(ana, "mv", "x.txt", "de-ana.txt")
        (ana / "otro.md").write_text("cambio de ana\n")
        self.git(luis, "mv", "x.txt", "de-luis.txt")
        (luis / "otro.md").write_text("cambio de luis\n")
        self.ok("luis", "guardar", "Fichas", "renombro")
        rc, datos, err = self.compartir("ana", "guardar", "Fichas", "renombro")
        self.assertEqual(rc, 0, err)
        self.assertNotIn("Traceback", err)
        self.assertFalse((ana / ".git" / "MERGE_HEAD").exists())
        for nombre in ("de-ana.txt", "de-luis.txt"):
            self.assertEqual((ana / nombre).read_text(), "contenido\n")
        for p in ana.rglob("*"):
            if p.is_file() and ".git" not in p.parts:
                self.assertNotIn("<<<<<<<", p.read_text(errors="ignore"))

    def test_mezcla_a_medias_se_recupera(self):
        ana, _ = self.espacio_compartido()
        (ana / "a.md").write_text("a\n")
        self.ok("ana", "guardar", "Fichas", "a")
        # Una mezcla interrumpida que quedó a medias
        (ana / ".git" / "MERGE_HEAD").write_text(self.git(ana, "rev-parse", "HEAD"))
        (ana / "b.md").write_text("b\n")
        datos = self.ok("ana", "guardar", "Fichas", "b")
        self.assertIn("b.md", datos["subidos"])

    def test_estado_pendientes(self):
        ana, _ = self.espacio_compartido()
        (ana / "nuevo.md").write_text("hola\n")
        datos = self.ok("ana", "estado")
        self.assertEqual(datos["usuario"], "ana")
        esp = {e["nombre"]: e for e in datos["espacios"]}
        self.assertEqual(esp["Fichas"]["pendientes"], 1)
        self.assertEqual(esp["Fichas"]["propietario"], "ana")


class TestAcceso(Base):
    def invitar_y_aceptar(self):
        self.ok("ana", "crear", "Fichas")
        self.ok("ana", "invitar", "Fichas", "luis")
        inv = self.ok("luis", "invitaciones")
        return self.ok("luis", "aceptar", str(inv[0]["id"]))

    def test_invitar_y_quien(self):
        self.ok("ana", "crear", "Fichas")
        datos = self.ok("ana", "invitar", "Fichas", "luis")
        self.assertEqual(datos, {"invitado": "luis", "estado": "pendiente"})
        quien = self.ok("ana", "quien", "Fichas")
        self.assertEqual(quien["propietario"], "ana")
        self.assertEqual(quien["pendientes"], ["luis"])
        self.assertEqual(quien["colaboradores"], [])

    def test_invitar_usuario_inexistente(self):
        self.ok("ana", "crear", "Fichas")
        rc, _, err = self.compartir("ana", "invitar", "Fichas", "nadie")
        self.assertEqual(rc, 1)
        self.assertIn("No hay ningún usuario de GitHub llamado nadie", err)
        self.assertFalse([a for a in self.llamadas() if "PUT" in a])

    def test_invitar_con_correo(self):
        self.ok("ana", "crear", "Fichas")
        antes = len(self.llamadas())
        rc, _, err = self.compartir("ana", "invitar", "Fichas", "luis@educa.madrid.org")
        self.assertEqual(rc, 1)
        self.assertIn("Necesito el nombre de usuario de GitHub, no el correo", err)
        self.assertEqual(len(self.llamadas()), antes)

    def test_invitar_sin_ser_propietario(self):
        self.invitar_y_aceptar()
        rc, _, err = self.compartir("luis", "invitar", "Fichas", "ana")
        self.assertEqual(rc, 1)
        self.assertIn("Solo ana puede invitar a este espacio", err)

    def test_quitar_colaborador_y_pendiente(self):
        self.invitar_y_aceptar()
        self.assertEqual(self.ok("ana", "quien", "Fichas")["colaboradores"], ["luis"])
        self.ok("ana", "quitar", "Fichas", "luis")
        self.assertEqual(self.ok("ana", "quien", "Fichas")["colaboradores"], [])
        # invitación pendiente
        self.ok("ana", "invitar", "Fichas", "luis")
        self.assertEqual(self.ok("ana", "quien", "Fichas")["pendientes"], ["luis"])
        self.ok("ana", "quitar", "Fichas", "luis")
        quien = self.ok("ana", "quien", "Fichas")
        self.assertEqual(quien["pendientes"], [])
        self.assertEqual(quien["colaboradores"], [])

    def test_aceptar_otra_vez_si_falta_la_carpeta(self):
        datos = self.invitar_y_aceptar()
        shutil.rmtree(self.raiz("luis") / datos["espacio"])
        datos = self.ok("luis", "aceptar", "ana/fichas")
        self.assertEqual(datos["espacio"], "Fichas")
        self.assertTrue((self.raiz("luis") / "Fichas" / ".gitignore").is_file())

    def test_aceptar_clona(self):
        self.ok("ana", "crear", "Fichas")
        self.ok("ana", "invitar", "Fichas", "luis")
        # Una invitación a un repositorio que no es de Mint Docente no aparece.
        estado = json.loads((self.tmp / "estado-gh" / "estado.json").read_text())
        estado["repos"]["ana/otro"] = {"topics": [], "colaboradores": [], "descripcion": "Otro"}
        estado["invitaciones"].append({"id": 99, "repo": "ana/otro", "invitado": "luis", "de": "ana"})
        (self.tmp / "estado-gh" / "estado.json").write_text(json.dumps(estado))
        inv = self.ok("luis", "invitaciones")
        self.assertEqual(len(inv), 1)
        self.assertEqual(inv[0]["espacio"], "Fichas")
        self.assertEqual(inv[0]["de"], "ana")
        datos = self.ok("luis", "aceptar", str(inv[0]["id"]))
        carpeta = self.raiz("luis") / datos["espacio"]
        self.assertTrue((carpeta / ".gitignore").is_file())
        self.assertIn("+luis@users.noreply.github.com", self.git(carpeta, "config", "user.email"))
        # Y puede guardar.
        (carpeta / "de-luis.md").write_text("hola\n")
        self.ok("luis", "guardar", datos["espacio"], "hola")
        self.ok("ana", "traer", "Fichas")
        self.assertTrue((self.raiz("ana") / "Fichas" / "de-luis.md").is_file())


SKILL = """---
name: rubricas-fp
description: Rúbricas para FP
---

Versión {}
"""


class TestSkills(Base):
    def test_usar_copia_fija_y_aviso_de_novedad(self):
        ana, luis = self.espacio_compartido()
        (ana / "skills" / "rubricas-fp").mkdir(parents=True)
        (ana / "skills" / "rubricas-fp" / "SKILL.md").write_text(SKILL.format(1))
        self.ok("ana", "guardar", "Fichas", "skill")
        self.ok("luis", "traer", "Fichas")
        datos = self.ok("luis", "usar", "Fichas", "skills/rubricas-fp")
        self.assertEqual(datos["skill"], "rubricas-fp")
        copia = self.skills("luis") / "rubricas-fp"
        self.assertFalse(copia.is_symlink())
        self.assertIn("Versión 1", (copia / "SKILL.md").read_text())
        self.assertTrue((copia / ".origen").is_file())
        # ana la cambia
        (ana / "skills" / "rubricas-fp" / "SKILL.md").write_text(SKILL.format(2))
        self.ok("ana", "guardar", "Fichas", "skill v2")
        datos = self.ok("luis", "traer", "Fichas")
        self.assertEqual(datos["espacios"][0]["skills_con_novedades"], ["rubricas-fp"])
        self.assertIn("Versión 1", (copia / "SKILL.md").read_text())
        lista = self.ok("luis", "skills")
        self.assertEqual([(s["skill"], s["espacio"], s["hay_nueva"]) for s in lista],
                         [("rubricas-fp", "Fichas", True)])
        self.ok("luis", "usar", "Fichas", "skills/rubricas-fp")
        self.assertIn("Versión 2", (copia / "SKILL.md").read_text())
        self.assertFalse(self.ok("luis", "skills")[0]["hay_nueva"])

    def test_usar_sin_frontmatter(self):
        ana, _ = self.espacio_compartido()
        (ana / "mala").mkdir()
        (ana / "mala" / "SKILL.md").write_text("Sin frontmatter\n")
        rc, _, err = self.compartir("ana", "usar", "Fichas", "mala")
        self.assertEqual(rc, 1)
        self.assertIn("no es una skill válida", err)
        self.assertFalse(self.skills("ana").exists() and any(self.skills("ana").iterdir()))

    def test_usar_nombre_de_skill_del_sistema(self):
        ana, _ = self.espacio_compartido()
        (ana / "falsa").mkdir()
        (ana / "falsa" / "SKILL.md").write_text(SKILL.replace("rubricas-fp", "compartir").format(1))
        rc, _, err = self.compartir("ana", "usar", "Fichas", "falsa")
        self.assertEqual(rc, 1)
        self.assertIn("Ya usas otra skill llamada compartir", err)
        self.assertFalse((self.skills("ana") / "compartir").exists())

    def test_usar_nombre_ocupado(self):
        ana, _ = self.espacio_compartido()
        (ana / "skills" / "rubricas-fp").mkdir(parents=True)
        (ana / "skills" / "rubricas-fp" / "SKILL.md").write_text(SKILL.format(1))
        propia = self.skills("ana") / "rubricas-fp"
        propia.mkdir(parents=True)
        (propia / "SKILL.md").write_text("mía\n")
        rc, _, err = self.compartir("ana", "usar", "Fichas", "skills/rubricas-fp")
        self.assertEqual(rc, 1)
        self.assertIn("Ya usas otra skill llamada rubricas-fp", err)
        self.assertEqual((propia / "SKILL.md").read_text(), "mía\n")


if __name__ == "__main__":
    unittest.main()

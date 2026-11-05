package registro

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbrirAnade(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "José Ángel", "MintDocente", "instalador.log")
	for i := 0; i < 2; i++ {
		l, c := Abrir(ruta)
		Func(l)("paso %d", i)
		c.Close()
	}
	b, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "paso 0") || !strings.Contains(string(b), "paso 1") {
		t.Errorf("registro = %q", b)
	}
}

func TestAbrirSinPermisoNoFalla(t *testing.T) {
	archivo := filepath.Join(t.TempDir(), "no-es-carpeta")
	os.WriteFile(archivo, nil, 0o644)
	l, c := Abrir(filepath.Join(archivo, "instalador.log"))
	Func(l)("nada")
	c.Close()
}

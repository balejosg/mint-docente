package unir

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func suma(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func crear(t *testing.T, dir string, contenidos ...string) []string {
	t.Helper()
	var rutas []string
	for i, c := range contenidos {
		r := filepath.Join(dir, "x.ova.part"+string(rune('1'+i)))
		if err := os.WriteFile(r, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		rutas = append(rutas, r)
	}
	return rutas
}

func TestUneEnOrden(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "uno-", "dos-", "tres")
	destino := filepath.Join(dir, "x.ova")
	if err := Partes(partes, destino, suma([]byte("uno-dos-tres"))); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(destino)
	if string(got) != "uno-dos-tres" {
		t.Fatalf("contenido: %q", got)
	}
}

func TestBorraPartes(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "a", "b")
	if err := Partes(partes, filepath.Join(dir, "x.ova"), suma([]byte("ab"))); err != nil {
		t.Fatal(err)
	}
	for _, p := range partes {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("la parte %s sigue ahí", p)
		}
	}
}

func TestSumaMal(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "a", "b")
	destino := filepath.Join(dir, "x.ova")
	err := Partes(partes, destino, suma([]byte("otra cosa")))
	if !errors.Is(err, ErrSuma) {
		t.Fatalf("esperaba ErrSuma, obtuve %v", err)
	}
	if _, e := os.Stat(destino); !os.IsNotExist(e) {
		t.Fatal("queda destino")
	}
	for _, p := range partes {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			t.Fatalf("queda %s", p)
		}
	}
}

func existe(p string) bool { _, err := os.Stat(p); return err == nil }

func TestParteAusenteConservaLasDemas(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "a", "b")
	partes = append(partes, filepath.Join(dir, "no-existe"))
	destino := filepath.Join(dir, "x.ova")
	err := Partes(partes, destino, suma([]byte("ab")))
	if !errors.Is(err, os.ErrNotExist) || errors.Is(err, ErrSuma) {
		t.Fatalf("error inesperado: %v", err)
	}
	if existe(destino) || !existe(partes[0]) || !existe(partes[1]) {
		t.Fatal("debe quedar sin destino y con las partes")
	}
}

func TestFalloAlEscribirConservaPartes(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "a", "b")
	// La segunda "parte" es un directorio: leerla falla a mitad de la unión.
	malo := filepath.Join(dir, "directorio")
	if err := os.Mkdir(malo, 0o755); err != nil {
		t.Fatal(err)
	}
	partes = append(partes[:1], malo, partes[1])
	destino := filepath.Join(dir, "x.ova")
	err := Partes(partes, destino, suma([]byte("ab")))
	if err == nil || errors.Is(err, ErrSuma) {
		t.Fatalf("error inesperado: %v", err)
	}
	if existe(destino) || !existe(partes[0]) || !existe(partes[2]) {
		t.Fatal("debe quedar sin destino y con las partes")
	}
}

func TestSumaEnMayusculas(t *testing.T) {
	dir := t.TempDir()
	partes := crear(t, dir, "a")
	if err := Partes(partes, filepath.Join(dir, "x.ova"), strings.ToUpper(suma([]byte("a")))); err != nil {
		t.Fatal(err)
	}
}

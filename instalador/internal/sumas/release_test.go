package sumas

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestPrepararRelease ejecuta preparar-release.sh sobre una OVA falsa de 5 MB
// con partes pequeñas y comprueba que su SHA256SUMS lo entiende Leer.
func TestPrepararRelease(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("preparar-release.sh se ejecuta en Linux")
	}
	script, err := filepath.Abs("../../../preparar-release.sh")
	if err != nil {
		t.Fatal(err)
	}
	salida := t.TempDir()
	ova := make([]byte, 5<<20)
	if _, err := rand.Read(ova); err != nil {
		t.Fatal(err)
	}
	const nombre = "MintDocente-20990101.ova"
	escribir(t, filepath.Join(salida, nombre), ova)
	exe := []byte("MZ instalador falso")
	escribir(t, filepath.Join(salida, "InstalarMintDocente.exe"), exe)

	// 5 MB en partes de 400 000 bytes: 14 partes, para probar part10 y siguientes.
	const tam = 400000
	cmd := exec.Command("bash", script, "v9.9")
	cmd.Env = append(os.Environ(), "SALIDA="+salida, fmt.Sprintf("TAM_PARTE=%d", tam))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("preparar-release.sh: %v\n%s", err, out)
	}
	rel := filepath.Join(salida, "release-v9.9")

	if _, err := os.Stat(filepath.Join(rel, nombre)); err == nil {
		t.Error("la OVA completa no debe ir a la release (GitHub no admite archivos de más de 2 GB)")
	}

	f, err := os.Open(filepath.Join(rel, "SHA256SUMS"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l, err := Leer(f)
	if err != nil {
		t.Fatal(err)
	}
	if want := (len(ova) + tam - 1) / tam; len(l.Partes) != want {
		t.Fatalf("partes: %d, esperaba %d", len(l.Partes), want)
	}
	var unidas []byte
	for i, p := range l.Partes {
		if want := fmt.Sprintf("%s.part%d", nombre, i); p.Nombre != want {
			t.Errorf("parte %d: %s, esperaba %s", i, p.Nombre, want)
		}
		datos, err := os.ReadFile(filepath.Join(rel, p.Nombre))
		if err != nil {
			t.Fatal(err)
		}
		if i < len(l.Partes)-1 && len(datos) != tam {
			t.Errorf("%s mide %d bytes", p.Nombre, len(datos))
		}
		unidas = append(unidas, datos...)
	}
	if !bytes.Equal(unidas, ova) {
		t.Error("las partes unidas no son la OVA original")
	}
	// Como hará quien la descargue: unir las partes y comprobar con SHA256SUMS.
	escribir(t, filepath.Join(rel, nombre), unidas)
	chk := exec.Command("sha256sum", "-c", "--strict", "SHA256SUMS")
	chk.Dir = rel
	if out, err := chk.CombinedOutput(); err != nil {
		t.Fatalf("sha256sum -c: %v\n%s", err, out)
	}
	suma := sha256.Sum256(ova)
	if l.Completa.Nombre != nombre || l.Completa.SHA256 != hex.EncodeToString(suma[:]) {
		t.Errorf("completa: %+v", l.Completa)
	}
	if got, err := os.ReadFile(filepath.Join(rel, "InstalarMintDocente.exe")); err != nil || !bytes.Equal(got, exe) {
		t.Errorf("InstalarMintDocente.exe no está en la release: %v", err)
	}
}

func escribir(t *testing.T, ruta string, datos []byte) {
	t.Helper()
	if err := os.WriteFile(ruta, datos, 0o644); err != nil {
		t.Fatal(err)
	}
}

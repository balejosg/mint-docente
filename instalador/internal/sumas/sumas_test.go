package sumas

import (
	"os"
	"strings"
	"testing"
)

func TestLeerSHA256SUMSv11(t *testing.T) {
	f, err := os.Open("testdata/SHA256SUMS")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l, err := Leer(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Partes) != 3 {
		t.Fatalf("partes: %d", len(l.Partes))
	}
	for i, p := range l.Partes {
		want := "MintDocente-20261105.ova.part" + string(rune('0'+i))
		if p.Nombre != want {
			t.Errorf("parte %d: %s", i, p.Nombre)
		}
	}
	if l.Completa.Nombre != "MintDocente-20261105.ova" ||
		l.Completa.SHA256 != "be7f0c0e6a6305ca83cc8a30acc27f87cdad0521288b38f656cf215a1a3fdcac" {
		t.Errorf("completa: %+v", l.Completa)
	}
}

func linea(h, n string) string { return strings.Repeat(h, 64) + "  " + n + "\n" }

func TestOrdenNumerico(t *testing.T) {
	in := linea("a", "X.ova.part10") + linea("b", "X.ova.part9") +
		linea("c", "X.ova.part2") + linea("d", "X.ova")
	l, err := Leer(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range l.Partes {
		got = append(got, p.Nombre)
	}
	want := "X.ova.part2 X.ova.part9 X.ova.part10"
	if strings.Join(got, " ") != want {
		t.Errorf("orden: %v", got)
	}
}

func TestSinCompleta(t *testing.T) {
	if _, err := Leer(strings.NewReader(linea("a", "X.ova.part0"))); err == nil {
		t.Error("esperaba error")
	}
}

func TestLineaMalFormada(t *testing.T) {
	in := linea("a", "X.ova.part0") + "esto no es una suma\n" + linea("d", "X.ova")
	if _, err := Leer(strings.NewReader(in)); err == nil {
		t.Error("esperaba error")
	}
}

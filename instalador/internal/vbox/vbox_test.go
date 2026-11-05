package vbox

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// falso registra los argumentos y responde según la primera orden.
type falso struct {
	llamadas [][]string
	salidas  map[string]string
}

func (f *falso) Ejecutar(args ...string) (string, error) {
	f.llamadas = append(f.llamadas, args)
	return f.salidas[strings.Join(args, " ")], nil
}

func TestImportarArgs(t *testing.T) {
	f := &falso{}
	carpeta := `C:\Users\José Ángel\VirtualBox VMs`
	if err := Importar(f, `C:\Users\José Ángel\mint docente.ova`, "Mint Docente", carpeta, 4096, 2); err != nil {
		t.Fatal(err)
	}
	quiero := []string{"import", `C:\Users\José Ángel\mint docente.ova`, "--vsys", "0",
		"--vmname", "Mint Docente", "--basefolder", carpeta, "--memory", "4096", "--cpus", "2"}
	if len(f.llamadas) != 1 || !reflect.DeepEqual(f.llamadas[0], quiero) {
		t.Fatalf("args = %q, quiero %q", f.llamadas, quiero)
	}
}

func TestVersion(t *testing.T) {
	casos := map[string]Version{
		"7.2.4r170995":      {7, 2, 4},
		"7.0":               {7, 0, 0},
		"6.1.50":            {6, 1, 50},
		"7.1.0_Ubuntur1234": {7, 1, 0},
		"7.1.2_BETA":        {7, 1, 2},
		"7.2.4r170995\r\n":  {7, 2, 4},
	}
	for s, quiero := range casos {
		f := &falso{salidas: map[string]string{"--version": s}}
		v, err := LeerVersion(f)
		if err != nil || v != quiero {
			t.Errorf("%q: %v %v, quiero %v", s, v, err, quiero)
		}
	}
	a, b := Version{7, 2, 4}, Version{7, 0, 0}
	if a.Menor(b) || !b.Menor(a) {
		t.Error("7.2.4 debe ser mayor que 7.0")
	}
	if !(Version{6, 1, 50}).Menor(b) {
		t.Error("6.1.50 debe ser menor que 7.0")
	}
	f := &falso{salidas: map[string]string{"--version": "basura"}}
	if _, err := LeerVersion(f); err == nil {
		t.Error("esperaba error con salida ilegible")
	}
}

func TestEstado(t *testing.T) {
	f := &falso{salidas: map[string]string{
		"list vms":        "\"Mint Docente\" {1234-uuid}\n\"Otra\" {abcd}\n",
		"list runningvms": "",
	}}
	existe, enc, err := Estado(f, "Mint Docente")
	if err != nil || !existe || enc {
		t.Fatalf("existe=%v encendida=%v err=%v", existe, enc, err)
	}
	f.salidas["list runningvms"] = "\"Mint Docente\" {1234-uuid}\n"
	if _, enc, _ = Estado(f, "Mint Docente"); !enc {
		t.Error("debería estar encendida")
	}
	if existe, _, _ = Estado(f, "Mint"); existe {
		t.Error("un nombre parcial no debe coincidir")
	}
}

func TestCompartirRutaRed(t *testing.T) {
	f := &falso{}
	ruta := `\\servidor\docs\Intercambio`
	if err := Compartir(f, "Mint Docente", ruta); err != nil {
		t.Fatal(err)
	}
	quiero := []string{"sharedfolder", "add", "Mint Docente", "--name", "Intercambio", "--hostpath", ruta, "--automount"}
	if !reflect.DeepEqual(f.llamadas[0], quiero) {
		t.Fatalf("args = %q", f.llamadas[0])
	}
}

func TestCarpetaMaquinas(t *testing.T) {
	existeRuta = func(string) bool { return true }
	defer func() { existeRuta = func(r string) bool { _, e := os.Stat(r); return e == nil } }()
	f := &falso{salidas: map[string]string{"list systemproperties": "API version: 7_1\r\nDefault machine folder:          C:\\Users\\José Ángel\\VirtualBox VMs\r\nVRDE: x\r\n"}}
	c, err := CarpetaMaquinas(f)
	if err != nil || c != `C:\Users\José Ángel\VirtualBox VMs` {
		t.Fatalf("%q %v", c, err)
	}
	f.salidas["list systemproperties"] = "nada\n"
	if _, err := CarpetaMaquinas(f); err == nil {
		t.Error("esperaba error")
	}
}

func TestArrancarYBorrar(t *testing.T) {
	f := &falso{}
	_ = Arrancar(f, "Mint Docente")
	_ = Borrar(f, "Mint Docente")
	quiero := [][]string{{"startvm", "Mint Docente", "--type", "gui"}, {"unregistervm", "Mint Docente", "--delete"}}
	if !reflect.DeepEqual(f.llamadas, quiero) {
		t.Fatalf("%q", f.llamadas)
	}
}

// cp1252 de «José Ángel»: é=0xE9, Á=0xC1.
const cp1252Carpeta = "C:\\Users\\Jos\xe9 \xc1ngel\\VirtualBox VMs"
const utf8Carpeta = `C:\Users\José Ángel\VirtualBox VMs`

func decCp1252(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		b.WriteRune(rune(s[i])) // latin1/cp1252 coinciden en é y Á
	}
	return b.String()
}

func TestElegirCarpetaCp1252(t *testing.T) {
	salida := "Default machine folder:   " + cp1252Carpeta + "\r\n"
	existe := func(r string) bool { return r == `C:\Users\José Ángel` }
	c, err := elegirCarpeta(salida, []func(string) string{decCp1252}, existe, `C:\x`)
	if err != nil || c != utf8Carpeta {
		t.Fatalf("%q %v", c, err)
	}
}

func TestElegirCarpetaUTF8(t *testing.T) {
	salida := "Default machine folder:   " + utf8Carpeta + "\r\n"
	malo := func(string) string { t.Error("no debe decodificar"); return "" }
	c, err := elegirCarpeta(salida, []func(string) string{malo}, func(string) bool { return true }, `C:\x`)
	if err != nil || c != utf8Carpeta {
		t.Fatalf("%q %v", c, err)
	}
}

func TestElegirCarpetaRespaldo(t *testing.T) {
	salida := "Default machine folder:   " + cp1252Carpeta + "\r\n"
	c, err := elegirCarpeta(salida, []func(string) string{decCp1252}, func(string) bool { return false }, `C:\Users\Ana`)
	if err != nil || c != `C:\Users\Ana\VirtualBox VMs` {
		t.Fatalf("%q %v", c, err)
	}
	if _, err := elegirCarpeta(salida, nil, func(string) bool { return false }, ""); err == nil {
		t.Error("sin perfil debe fallar")
	}
}

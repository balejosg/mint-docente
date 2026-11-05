package flujo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/balejosg/mint-docente/instalador/internal/descarga"
	"github.com/balejosg/mint-docente/instalador/internal/equipo"
	"github.com/balejosg/mint-docente/instalador/internal/vbox"
)

const (
	ova   = "MintDocente-20261105.ova"
	parte = ova + ".part"
	uuid  = "{0b1c5d2e-0000-4000-8000-000000000001}"
)

// ---------- VirtualBox falso ----------

type vboxFalso struct {
	version        string
	vms, running   string
	carpeta        string
	fallaImport    bool
	fallaCompartir bool
	ordenes        [][]string
}

func (v *vboxFalso) Ejecutar(args ...string) (string, error) {
	v.ordenes = append(v.ordenes, args)
	switch args[0] {
	case "--version":
		return v.version, nil
	case "list":
		switch args[1] {
		case "vms":
			return v.vms, nil
		case "runningvms":
			return v.running, nil
		case "systemproperties":
			return "Default machine folder:          " + v.carpeta + "\n", nil
		}
	case "import":
		if _, err := os.Stat(args[1]); err != nil {
			return "", fmt.Errorf("no existe la OVA: %v", err)
		}
		if v.fallaImport {
			return "", errors.New("VBOX_E_FILE_ERROR")
		}
		v.vms = `"Mint Docente" ` + uuid + "\n"
		return "", nil
	case "sharedfolder":
		if v.fallaCompartir {
			return "", errors.New("VBOX_E_OBJECT_NOT_FOUND")
		}
		return "", nil
	case "unregistervm":
		v.vms = ""
		return "", nil
	case "startvm":
		return "", nil
	}
	return "", fmt.Errorf("orden inesperada %v", args)
}

// acciones devuelve las órdenes que cambian algo (sin list ni --version).
func (v *vboxFalso) acciones() []string {
	var r []string
	for _, o := range v.ordenes {
		if o[0] != "list" && o[0] != "--version" {
			r = append(r, o[0])
		}
	}
	return r
}

func (v *vboxFalso) orden(nombre string) []string {
	for _, o := range v.ordenes {
		if o[0] == nombre {
			return o
		}
	}
	return nil
}

// ---------- Ventanas falsas ----------

type barraFalsa struct{ cerrada bool }

func (b *barraFalsa) Actualizar(string, float64) {}
func (b *barraFalsa) Cerrar()                    { b.cerrada = true }

type uiFalsa struct {
	avisos, errores, preguntas []string
	respuestas                 []bool
}

func (u *uiFalsa) Aviso(t string) { u.avisos = append(u.avisos, t) }
func (u *uiFalsa) Error(t string) { u.errores = append(u.errores, t) }
func (u *uiFalsa) Pregunta(t, si, no string) bool {
	u.preguntas = append(u.preguntas, t+" ["+si+"/"+no+"]")
	if len(u.respuestas) == 0 {
		return false
	}
	r := u.respuestas[0]
	u.respuestas = u.respuestas[1:]
	return r
}
func (u *uiFalsa) Progreso(string) Barra { return &barraFalsa{} }

// ---------- Servidor de la release ----------

type origen struct {
	srv      *httptest.Server
	archivos map[string][]byte
	mu       sync.Mutex
	// peticiones: «nombre rango» de cada GET que no es una sonda de tamaño.
	peticiones []string
	// cortar decide, para cada GET de descarga, cuántos bytes enviar antes de
	// cortar (0: no cortar; -1: cortar sin enviar nada).
	cortar func(nombre string, n int) int
	cuenta map[string]int
}

func suma(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func nuevoOrigen(t *testing.T) *origen {
	t.Helper()
	p1 := bytes.Repeat([]byte("parte-uno-"), 20000)
	p2 := bytes.Repeat([]byte("parte-dos-"), 15000)
	completa := append(append([]byte{}, p1...), p2...)
	sums := fmt.Sprintf("%s *%s1\n%s *%s2\n%s *%s\n", suma(p1), parte, suma(p2), parte, suma(completa), ova)
	o := &origen{
		archivos: map[string][]byte{"SHA256SUMS": []byte(sums), parte + "1": p1, parte + "2": p2},
		cuenta:   map[string]int{},
	}
	o.srv = httptest.NewServer(http.HandlerFunc(o.servir))
	t.Cleanup(o.srv.Close)
	return o
}

func (o *origen) tam() int64 { return int64(len(o.archivos[parte+"1"]) + len(o.archivos[parte+"2"])) }

func (o *origen) servir(w http.ResponseWriter, r *http.Request) {
	nombre := strings.TrimPrefix(r.URL.Path, "/")
	b, ok := o.archivos[nombre]
	if !ok {
		http.NotFound(w, r)
		return
	}
	rango := r.Header.Get("Range")
	if rango == "bytes=0-0" {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-0/%d", len(b)))
		w.Header().Set("Content-Length", "1")
		w.WriteHeader(http.StatusPartialContent)
		w.Write(b[:1])
		return
	}
	o.mu.Lock()
	o.peticiones = append(o.peticiones, nombre+" "+rango)
	o.cuenta[nombre]++
	n := o.cuenta[nombre]
	corte := 0
	if o.cortar != nil {
		corte = o.cortar(nombre, n)
	}
	o.mu.Unlock()
	if corte < 0 {
		panic(http.ErrAbortHandler)
	}
	inicio := 0
	if rango != "" {
		fmt.Sscanf(rango, "bytes=%d-", &inicio)
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", inicio, len(b)-1, len(b)))
		w.Header().Set("Content-Length", strconv.Itoa(len(b)-inicio))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	}
	if corte > 0 {
		w.Write(b[inicio : inicio+corte])
		w.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	}
	w.Write(b[inicio:])
}

func (o *origen) descargas(nombre string) []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	var r []string
	for _, p := range o.peticiones {
		if strings.HasPrefix(p, nombre+" ") {
			r = append(r, p)
		}
	}
	return r
}

// ---------- Escenario ----------

type escenario struct {
	dep      Dependencias
	ui       *uiFalsa
	vb       *vboxFalso
	o        *origen
	vboxDir  string
	maquinas string
	docs     string
	libre    int64
	accesos  [][2]string
	pausas   []time.Duration
}

func nuevo(t *testing.T) *escenario {
	t.Helper()
	raiz := t.TempDir()
	e := &escenario{
		ui:       &uiFalsa{},
		o:        nuevoOrigen(t),
		vboxDir:  filepath.Join(raiz, "Oracle VirtualBox"),
		maquinas: filepath.Join(raiz, "José Ángel", "VirtualBox VMs"),
		docs:     filepath.Join(raiz, "José Ángel", "Documentos"),
		libre:    1 << 50,
	}
	os.MkdirAll(filepath.Dir(e.maquinas), 0o755)
	e.vb = &vboxFalso{version: "7.2.4r170995\n", carpeta: e.maquinas}
	e.dep = Dependencias{
		UI:             e.ui,
		Localizar:      func() (string, error) { return filepath.Join(e.vboxDir, "VBoxManage.exe"), nil },
		NuevoEjecutor:  func(string) vbox.Ejecutor { return e.vb },
		Cliente:        e.o.srv.Client(),
		URLBase:        e.o.srv.URL + "/",
		Trabajo:        filepath.Join(raiz, "Local", "MintDocente", "descarga"),
		Registro:       filepath.Join(raiz, "Local", "MintDocente", "instalador.log"),
		Log:            func(f string, a ...any) { t.Logf("registro: "+f, a...) },
		MemoriaTotalMB: func() (int, error) { return 16384, nil },
		Nucleos:        8,
		EspacioLibre:   func(string) (int64, error) { return e.libre, nil },
		Documentos:     func() (string, error) { return e.docs, nil },
		AccesoDirecto: func(destino, args string) error {
			e.accesos = append(e.accesos, [2]string{destino, args})
			return nil
		},
		Esperar: func(d time.Duration) { e.pausas = append(e.pausas, d) },
	}
	return e
}

func (e *escenario) ejecutar(t *testing.T) error {
	t.Helper()
	return Ejecutar(context.Background(), e.dep)
}

func (e *escenario) unError(t *testing.T) string {
	t.Helper()
	if len(e.ui.errores) != 1 {
		t.Fatalf("errores = %q, quería uno", e.ui.errores)
	}
	m := e.ui.errores[0]
	if !strings.Contains(m, "Si necesitas ayuda, envía el archivo instalador.log de la carpeta MintDocente.") ||
		!strings.Contains(m, e.dep.Registro) {
		t.Errorf("al error le falta la ayuda con la ruta del registro: %q", m)
	}
	return m
}

func noHay(t *testing.T, v *vboxFalso, ordenes ...string) {
	t.Helper()
	for _, o := range ordenes {
		if v.orden(o) != nil {
			t.Errorf("no debía llamarse a %s; órdenes: %v", o, v.ordenes)
		}
	}
}

// ---------- Pruebas ----------

func TestSinVirtualBox(t *testing.T) {
	e := nuevo(t)
	e.dep.Localizar = func() (string, error) { return "", vbox.ErrNoInstalado }
	if err := e.ejecutar(t); err == nil {
		t.Error("debía devolver error")
	}
	m := e.unError(t)
	if !strings.HasPrefix(m, "Este equipo no tiene VirtualBox. Pídeselo al coordinador TIC.") {
		t.Errorf("mensaje = %q", m)
	}
	noHay(t, e.vb, "import")
	if len(e.vb.ordenes) != 0 {
		t.Errorf("órdenes = %v", e.vb.ordenes)
	}
}

func TestVirtualBoxAntiguo(t *testing.T) {
	e := nuevo(t)
	e.vb.version = "6.1.50r161033\n"
	e.ejecutar(t)
	m := e.unError(t)
	if !strings.HasPrefix(m, "Este VirtualBox es antiguo; pide al coordinador TIC la versión 7.x") {
		t.Errorf("mensaje = %q", m)
	}
	noHay(t, e.vb, "import", "unregistervm")
	if len(e.o.peticiones) != 0 {
		t.Errorf("no debía descargar: %v", e.o.peticiones)
	}
}

func TestMaquinaEncendida(t *testing.T) {
	e := nuevo(t)
	e.vb.vms = `"Mint Docente" ` + uuid + "\n"
	e.vb.running = e.vb.vms
	e.ui.respuestas = []bool{false, true} // aunque contestara «Reinstalar» y «sí»
	e.ejecutar(t)
	want := "Mint Docente está abierta. Ciérrala y vuelve a abrir el instalador."
	if !slices.Contains(e.ui.avisos, want) {
		t.Errorf("avisos = %q", e.ui.avisos)
	}
	if len(e.ui.preguntas) != 0 {
		t.Errorf("no debía preguntar: %q", e.ui.preguntas)
	}
	noHay(t, e.vb, "unregistervm", "import", "startvm")
}

func TestAbrirExistente(t *testing.T) {
	e := nuevo(t)
	e.vb.vms = `"Mint Docente" ` + uuid + "\n"
	e.ui.respuestas = []bool{true} // «Abrir»
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err)
	}
	if got := e.vb.acciones(); !slices.Equal(got, []string{"startvm"}) {
		t.Errorf("acciones = %v", got)
	}
	if len(e.ui.preguntas) != 1 || !strings.Contains(e.ui.preguntas[0], "[Abrir/Reinstalar]") {
		t.Errorf("pregunta = %q", e.ui.preguntas)
	}
	if len(e.o.peticiones) != 0 {
		t.Errorf("no debía descargar: %v", e.o.peticiones)
	}
}

func TestReinstalarPideDosVeces(t *testing.T) {
	e := nuevo(t)
	e.vb.vms = `"Mint Docente" ` + uuid + "\n"
	e.ui.respuestas = []bool{false, false} // «Reinstalar» y después «no»
	e.ejecutar(t)
	if len(e.ui.preguntas) != 2 {
		t.Fatalf("preguntas = %q", e.ui.preguntas)
	}
	if !strings.Contains(e.ui.preguntas[0], "se borrará") {
		t.Errorf("la primera pregunta no avisa de que se borra: %q", e.ui.preguntas[0])
	}
	noHay(t, e.vb, "unregistervm", "import")
	if len(e.o.peticiones) != 0 {
		t.Errorf("no debía descargar: %v", e.o.peticiones)
	}
}

func TestReinstalarConfirmado(t *testing.T) {
	e := nuevo(t)
	e.vb.vms = `"Mint Docente" ` + uuid + "\n"
	e.ui.respuestas = []bool{false, true}
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err, e.ui.errores)
	}
	want := []string{"unregistervm", "import", "sharedfolder", "startvm"}
	if got := e.vb.acciones(); !slices.Equal(got, want) {
		t.Errorf("acciones = %v, quería %v", got, want)
	}
}

func TestSinEspacio(t *testing.T) {
	e := nuevo(t)
	necesario := equipo.EspacioNecesario(e.o.tam())
	e.libre = necesario - 10<<30 + 1 // faltan algo menos de 10 GiB
	e.ejecutar(t)
	m := e.unError(t)
	if !strings.Contains(m, "Faltan 10 GB libres en el disco") {
		t.Errorf("mensaje = %q", m)
	}
	if len(e.o.peticiones) != 1 || e.o.peticiones[0] != "SHA256SUMS " {
		t.Errorf("no debía descargar las partes: %v", e.o.peticiones)
	}
	noHay(t, e.vb, "import")
}

func TestImportFallaLimpia(t *testing.T) {
	e := nuevo(t)
	e.vb.fallaImport = true
	if err := e.ejecutar(t); err == nil {
		t.Error("debía devolver error")
	}
	e.unError(t)
	want := []string{"import", "unregistervm"}
	if got := e.vb.acciones(); !slices.Equal(got, want) {
		t.Errorf("acciones = %v, quería %v", got, want)
	}
	if o := e.vb.orden("unregistervm"); !slices.Equal(o, []string{"unregistervm", "Mint Docente", "--delete"}) {
		t.Errorf("unregistervm = %v", o)
	}
	// La OVA ya comprobada se conserva para no descargarla otra vez.
	if _, err := os.Stat(filepath.Join(e.dep.Trabajo, ova)); err != nil {
		t.Errorf("la OVA debía conservarse: %v", err)
	}
	e.vb.fallaImport = false
	e.ui.errores = nil
	antes := len(e.o.peticiones)
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err, e.ui.errores)
	}
	if len(e.o.peticiones) != antes+1 { // solo SHA256SUMS
		t.Errorf("el segundo intento volvió a descargar: %v", e.o.peticiones[antes:])
	}
}

func TestIntercambioFallaSigue(t *testing.T) {
	e := nuevo(t)
	e.vb.fallaCompartir = true
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err)
	}
	if e.vb.orden("startvm") == nil {
		t.Error("no arrancó la máquina")
	}
	if len(e.ui.avisos) != 1 || !strings.Contains(e.ui.avisos[0], "Intercambio") ||
		!strings.Contains(e.ui.avisos[0], "arrástralos") {
		t.Errorf("avisos = %q", e.ui.avisos)
	}
	if len(e.ui.errores) != 0 {
		t.Errorf("errores = %q", e.ui.errores)
	}
}

func TestCaminoFeliz(t *testing.T) {
	e := nuevo(t)
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err, e.ui.errores)
	}
	want := []string{"import", "sharedfolder", "startvm"}
	if got := e.vb.acciones(); !slices.Equal(got, want) {
		t.Errorf("acciones = %v, quería %v", got, want)
	}
	imp := e.vb.orden("import")
	wantImp := []string{"import", filepath.Join(e.dep.Trabajo, ova), "--vsys", "0", "--vmname", "Mint Docente",
		"--basefolder", e.maquinas, "--memory", "4096", "--cpus", "4"}
	if !slices.Equal(imp, wantImp) {
		t.Errorf("import = %q\nquería   %q", imp, wantImp)
	}
	inter := filepath.Join(e.docs, "Intercambio")
	if sf := e.vb.orden("sharedfolder"); len(sf) < 7 || sf[5] != "--hostpath" || sf[6] != inter {
		t.Errorf("sharedfolder = %q", sf)
	}
	if st, err := os.Stat(inter); err != nil || !st.IsDir() {
		t.Errorf("no se creó Intercambio: %v", err)
	}
	wantAcc := [2]string{filepath.Join(e.vboxDir, "VirtualBoxVM.exe"), `--startvm "Mint Docente"`}
	if len(e.accesos) != 1 || e.accesos[0] != wantAcc {
		t.Errorf("accesos = %q", e.accesos)
	}
	if _, err := os.Stat(e.dep.Trabajo); !os.IsNotExist(err) {
		t.Errorf("la carpeta de descarga no desapareció: %v", err)
	}
	if len(e.ui.avisos)+len(e.ui.errores) != 0 {
		t.Errorf("avisos %q errores %q", e.ui.avisos, e.ui.errores)
	}
}

func TestPocaMemoriaAvisaUnaVez(t *testing.T) {
	e := nuevo(t)
	e.dep.MemoriaTotalMB = func() (int, error) { return 4096, nil }
	e.dep.Nucleos = 2
	if err := e.ejecutar(t); err != nil {
		t.Fatal(err)
	}
	if len(e.ui.avisos) != 1 || !strings.Contains(e.ui.avisos[0], "poca memoria") {
		t.Errorf("avisos = %q", e.ui.avisos)
	}
	imp := e.vb.orden("import")
	if len(imp) < 12 || imp[9] != "2048" || imp[11] != "1" {
		t.Errorf("import = %q", imp)
	}
}

func TestDescargaCortadaSeReanuda(t *testing.T) {
	t.Run("en la misma ejecución", func(t *testing.T) {
		e := nuevo(t)
		e.o.cortar = func(nombre string, n int) int {
			if nombre == parte+"1" && n == 1 {
				return 50000
			}
			return 0
		}
		if err := e.ejecutar(t); err != nil {
			t.Fatal(err, e.ui.errores)
		}
		got := e.o.descargas(parte + "1")
		if !slices.Equal(got, []string{parte + "1 ", parte + "1 bytes=50000-"}) {
			t.Errorf("peticiones de la parte 1 = %q", got)
		}
		if e.vb.orden("startvm") == nil {
			t.Error("no llegó a arrancar")
		}
	})
	t.Run("en una segunda ejecución", func(t *testing.T) {
		e := nuevo(t)
		e.o.cortar = func(nombre string, n int) int {
			switch {
			case nombre != parte+"1":
				return 0
			case n == 1:
				return 50000
			default:
				return -1 // la red se cae del todo
			}
		}
		if err := e.ejecutar(t); err == nil {
			t.Fatal("la primera ejecución debía fallar")
		}
		m := e.unError(t)
		if !strings.Contains(m, "conexión") {
			t.Errorf("mensaje = %q", m)
		}
		// 5 intentos = 4 pausas. (El servidor puede ver alguna petición más:
		// net/http repite sola un GET si la conexión reutilizada se cierra.)
		if n := len(e.o.descargas(parte + "1")); n < 5 {
			t.Errorf("peticiones = %d, quería al menos 5", n)
		}
		if len(e.pausas) != 4 || !slices.IsSorted(e.pausas) || e.pausas[0] >= e.pausas[3] {
			t.Errorf("pausas = %v", e.pausas)
		}
		st, err := os.Stat(filepath.Join(e.dep.Trabajo, parte+"1.descargando"))
		if err != nil || st.Size() != 50000 {
			t.Fatalf("parcial: %v %v", st, err)
		}
		noHay(t, e.vb, "import")

		e.o.cortar = nil
		e.ui.errores = nil
		antes := len(e.o.descargas(parte + "1"))
		if err := e.ejecutar(t); err != nil {
			t.Fatal(err, e.ui.errores)
		}
		got := e.o.descargas(parte + "1")[antes:]
		if !slices.Equal(got, []string{parte + "1 bytes=50000-"}) {
			t.Errorf("la segunda ejecución no reanudó: %q", got)
		}
		if e.vb.orden("startvm") == nil {
			t.Error("no llegó a arrancar")
		}
	})
}

func TestDiscoLlenoDuranteLaDescarga(t *testing.T) {
	e := nuevo(t)
	original := bajar
	t.Cleanup(func() { bajar = original })
	llamadas := 0
	bajar = func(context.Context, *http.Client, string, string, string, func(int64, int64)) error {
		llamadas++
		return fmt.Errorf("%w: write: no space left on device", descarga.ErrSinEspacio)
	}
	if err := e.ejecutar(t); err == nil {
		t.Fatal("debía fallar")
	}
	m := e.unError(t)
	if !strings.Contains(m, "lleno") {
		t.Errorf("mensaje = %q", m)
	}
	if llamadas != 1 {
		t.Errorf("con el disco lleno no se reintenta: %d llamadas", llamadas)
	}
	noHay(t, e.vb, "import")
}

func TestDiscoLlenoDiceCuantoFalta(t *testing.T) {
	e := nuevo(t)
	original := bajar
	t.Cleanup(func() { bajar = original })
	primera := true
	bajar = func(context.Context, *http.Client, string, string, string, func(int64, int64)) error {
		if primera { // tras la comprobación inicial, el disco se ha ido llenando
			primera = false
			e.libre = 3 << 30
		}
		return descarga.ErrSinEspacio
	}
	e.ejecutar(t)
	m := e.unError(t)
	falta := equipo.EspacioNecesario(e.o.tam()) - 3<<30
	want := fmt.Sprintf("Faltan %d GB libres en el disco", (falta+1<<30-1)>>30)
	if !strings.Contains(m, want) {
		t.Errorf("mensaje = %q, quería %q", m, want)
	}
}

func TestSumaCompletaMalVuelveADescargarUnaVez(t *testing.T) {
	e := nuevo(t)
	s := string(e.o.archivos["SHA256SUMS"])
	s = s[:strings.LastIndex(s, "\n")]
	s = s[:strings.LastIndex(s, "\n")+1] + strings.Repeat("0", 64) + " *" + ova + "\n"
	e.o.archivos["SHA256SUMS"] = []byte(s)
	if err := e.ejecutar(t); err == nil {
		t.Fatal("debía fallar")
	}
	e.unError(t)
	if n := len(e.o.descargas(parte + "1")); n != 2 {
		t.Errorf("descargas de la parte 1 = %d, quería 2", n)
	}
	noHay(t, e.vb, "import")
}

func TestNombreInvalidoEnSHA256SUMS(t *testing.T) {
	for _, malo := range []string{"../x.ova.part1", `..\x.ova.part1`, "sub/x.ova.part1", `C:x.ova.part1`} {
		t.Run(malo, func(t *testing.T) {
			e := nuevo(t)
			s := string(e.o.archivos["SHA256SUMS"])
			e.o.archivos["SHA256SUMS"] = []byte(strings.Replace(s, parte+"1", malo, 1))
			e.ejecutar(t)
			e.unError(t)
			if len(e.o.peticiones) != 1 {
				t.Errorf("no debía descargar: %v", e.o.peticiones)
			}
		})
	}
}

func TestParteDanadaSeRepiteSoloUnaVez(t *testing.T) {
	e := nuevo(t)
	original := bajar
	t.Cleanup(func() { bajar = original })
	llamadas := 0
	bajar = func(context.Context, *http.Client, string, string, string, func(int64, int64)) error {
		llamadas++
		return descarga.ErrSuma
	}
	if err := e.ejecutar(t); err == nil {
		t.Fatal("debía fallar")
	}
	if m := e.unError(t); !strings.Contains(m, "dañada") {
		t.Errorf("mensaje = %q", m)
	}
	if llamadas != 2 {
		t.Errorf("descargas de la parte dañada = %d, quería 2", llamadas)
	}
}

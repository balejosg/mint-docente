// Package flujo encadena los pasos del instalador: comprueba VirtualBox y el
// espacio, descarga y une la OVA, la importa, comparte Intercambio, deja el
// acceso directo y arranca la máquina. Todo lo que toca fuera (ventanas,
// VirtualBox, red, disco) llega en Dependencias, para poder probarlo.
package flujo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/balejosg/mint-docente/instalador/internal/descarga"
	"github.com/balejosg/mint-docente/instalador/internal/equipo"
	"github.com/balejosg/mint-docente/instalador/internal/sumas"
	"github.com/balejosg/mint-docente/instalador/internal/unir"
	"github.com/balejosg/mint-docente/instalador/internal/vbox"
)

// Nombre es el nombre de la máquina en VirtualBox.
const Nombre = "Mint Docente"

// VBOX_MINIMA es la versión más antigua de VirtualBox que se admite.
// Provisional: la OVA se exporta con VirtualBox 7.2 y no se ha podido probar
// aquí la importación en un 7.0; la prueba del usuario en Windows lo confirma.
var VBOX_MINIMA = vbox.Version{Mayor: 7, Secundaria: 0, Parche: 0}

const (
	intentosDescarga = 5
	primeraPausa     = 2 * time.Second
	// tamEstimado se usa para el espacio si no se puede saber el tamaño de la OVA.
	tamEstimado = int64(4) << 30
	gib         = int64(1) << 30
)

// Mensajes que ve el docente.
const (
	msgSinVBox      = "Este equipo no tiene VirtualBox. Pídeselo al coordinador TIC."
	msgAntiguo      = "Este VirtualBox es antiguo; pide al coordinador TIC la versión 7.x."
	msgVBoxMudo     = "VirtualBox no responde. Reinicia el equipo y vuelve a abrir el instalador. Si sigue igual, pide ayuda al coordinador TIC."
	msgAbierta      = "Mint Docente está abierta. Ciérrala y vuelve a abrir el instalador."
	msgYaExiste     = "Ya tienes Mint Docente en este equipo. ¿Quieres abrirla o reinstalarla?\n\nSi la reinstalas, se borrará todo lo que tengas dentro y no hayas guardado fuera de ella."
	msgConfirmar    = "¿Seguro que quieres reinstalar Mint Docente?\n\nCuando la nueva esté descargada, se borrará la máquina actual con todo lo que tenga dentro. No se puede deshacer."
	msgEspacioVieja = "No hay espacio suficiente mientras siga en el equipo la Mint Docente anterior."
	msgNoBorra      = "No se ha podido borrar la Mint Docente anterior. Reinicia el equipo y vuelve a abrir el instalador."
	msgSinConexion  = "No se ha podido conectar con el sitio de descarga. Comprueba que hay conexión a internet y vuelve a abrir el instalador."
	msgDescargaMal  = "La descarga de Mint Docente no es válida. Vuelve a abrir el instalador más tarde."
	msgCortada      = "La descarga se ha cortado. Comprueba la conexión a internet y vuelve a abrir el instalador: seguirá donde lo dejó."
	msgDanada       = "La descarga ha llegado dañada varias veces. Vuelve a abrir el instalador más tarde o conéctate a otra red."
	msgNoUne        = "No se ha podido preparar la descarga. Vuelve a abrir el instalador: lo descargado no se pierde."
	msgPocaMemoria  = "Este equipo tiene poca memoria: Mint Docente funcionará, pero irá lenta. Cierra los demás programas mientras la uses."
	msgNoImporta    = "No se ha podido instalar Mint Docente en VirtualBox. Reinicia el equipo y vuelve a abrir el instalador: no tendrás que esperar otra vez a la descarga."
	msgSinInterc    = "No se ha podido preparar la carpeta Intercambio. Para pasar archivos entre Windows y Mint Docente, arrástralos a la ventana de la máquina."
	msgSinAcceso    = "No se ha podido crear el acceso directo «Mint Docente» en el escritorio. Para abrirla otra vez, abre VirtualBox y haz doble clic en «Mint Docente»."
	msgNoArranca    = "Mint Docente está instalada, pero no ha podido arrancar. Ábrela con el acceso directo «Mint Docente» del escritorio."
	msgDetenido     = "Has detenido la instalación. Cuando vuelvas a abrir el instalador, seguirá donde lo dejó."
	msgAyuda        = "Si necesitas ayuda, envía el archivo instalador.log de la carpeta MintDocente."
)

// Barra es una ventana de progreso.
type Barra interface {
	Actualizar(texto string, fraccion float64)
	Cerrar()
}

// UI son las ventanas que ve el docente.
type UI interface {
	Aviso(texto string)
	Error(texto string)
	// Pregunta devuelve true si se elige el botón si.
	Pregunta(texto, si, no string) bool
	Progreso(titulo string) Barra
}

// Dependencias reúne todo lo que el flujo usa del exterior.
type Dependencias struct {
	UI             UI
	Localizar      func() (string, error)                    // ruta de VBoxManage.exe
	NuevoEjecutor  func(rutaVBoxManage string) vbox.Ejecutor // ejecuta VBoxManage
	Cliente        *http.Client
	URLBase        string // termina en «/»
	Trabajo        string // carpeta de descarga, se borra al terminar
	Registro       string // ruta de instalador.log, para los mensajes de error
	Log            func(formato string, args ...any)
	MemoriaTotalMB func() (int, error)
	Nucleos        int
	EspacioLibre   func(ruta string) (int64, error)
	Documentos     func() (string, error)
	AccesoDirecto  func(destino, argumentos string) error // en el escritorio
	Esperar        func(time.Duration)                    // pausa entre reintentos
}

// bajar se sustituye en las pruebas.
var bajar = descarga.Bajar

type ejecucion struct {
	ctx      context.Context
	dep      Dependencias
	e        vbox.Ejecutor
	rutaVBox string
	barra    Barra
	ultimo   int // última fracción mostrada, en milésimas
	// reinstalar: hay una máquina anterior que se borra justo antes de
	// importar, cuando la OVA nueva ya está descargada y comprobada.
	reinstalar bool
}

// Ejecutar hace la instalación completa. Cada fallo ya se ha explicado al
// docente en una ventana cuando devuelve error; el detalle va al registro.
func Ejecutar(ctx context.Context, dep Dependencias) error {
	if dep.Log == nil {
		dep.Log = func(string, ...any) {}
	}
	if dep.Esperar == nil {
		dep.Esperar = time.Sleep
	}
	x := &ejecucion{ctx: ctx, dep: dep}
	defer x.cerrarBarra()
	return x.correr()
}

func (x *ejecucion) log(f string, a ...any) { x.dep.Log(f, a...) }

func (x *ejecucion) aviso(t string) {
	x.log("aviso: %s", t)
	x.dep.UI.Aviso(t)
}

// fallo explica el error al docente y lo devuelve con el detalle técnico.
func (x *ejecucion) fallo(texto string, detalle error) error {
	x.cerrarBarra()
	if detalle == nil {
		detalle = errors.New("sin detalle")
	}
	x.log("error: %s (%v)", texto, detalle)
	x.dep.UI.Error(ConAyuda(texto, x.dep.Registro))
	return fmt.Errorf("%s: %w", texto, detalle)
}

// ConAyuda añade a un mensaje de error a quién pedir ayuda y con qué archivo.
func ConAyuda(texto, registro string) string {
	return texto + "\n\n" + msgAyuda + "\n" + registro
}

func (x *ejecucion) detenido() error {
	x.cerrarBarra()
	x.aviso(msgDetenido)
	return x.ctx.Err()
}

func (x *ejecucion) mostrar(texto string, fraccion float64) {
	if x.barra == nil {
		x.barra = x.dep.UI.Progreso("Instalar Mint Docente")
	}
	x.barra.Actualizar(texto, fraccion)
}

func (x *ejecucion) cerrarBarra() {
	if x.barra != nil {
		x.barra.Cerrar()
		x.barra = nil
	}
}

func (x *ejecucion) correr() error {
	ruta, err := x.dep.Localizar()
	if err != nil {
		return x.fallo(msgSinVBox, err)
	}
	x.log("VBoxManage: %s", ruta)
	x.rutaVBox = ruta
	x.e = x.dep.NuevoEjecutor(ruta)

	v, err := vbox.LeerVersion(x.e)
	if err != nil {
		return x.fallo(msgVBoxMudo, err)
	}
	x.log("VirtualBox %s", v)
	if v.Menor(VBOX_MINIMA) {
		return x.fallo(msgAntiguo, fmt.Errorf("versión %s, mínima %s", v, VBOX_MINIMA))
	}

	existe, encendida, err := vbox.Estado(x.e, Nombre)
	if err != nil {
		return x.fallo(msgVBoxMudo, err)
	}
	if encendida {
		x.aviso(msgAbierta)
		return nil
	}
	if existe {
		// Cerrar la ventana cuenta como «no»: aquí, abrir (lo que no borra nada).
		if !x.dep.UI.Pregunta(msgYaExiste, "Reinstalar", "Abrir") {
			x.log("el docente elige abrir la máquina existente")
			return x.abrir()
		}
		if !x.dep.UI.Pregunta(msgConfirmar, "Sí, reinstalar", "No") {
			x.log("el docente no confirma la reinstalación")
			return nil
		}
		x.log("reinstalación confirmada: la máquina anterior se borrará antes de importar")
		x.reinstalar = true
	}
	return x.instalar()
}

func (x *ejecucion) abrir() error {
	x.acceso()
	if err := vbox.Arrancar(x.e, Nombre); err != nil {
		return x.fallo(msgNoArranca, err)
	}
	return nil
}

func (x *ejecucion) instalar() error {
	x.mostrar("Preparando la descarga…", 0)
	carpeta, err := vbox.CarpetaMaquinas(x.e)
	if err != nil {
		return x.fallo(msgVBoxMudo, err)
	}
	x.log("carpeta de máquinas: %s", carpeta)

	lista, err := x.leerSumas()
	if err != nil {
		return x.fallo(msgSinConexion, err)
	}
	if err := validar(lista); err != nil {
		return x.fallo(msgDescargaMal, err)
	}
	tams, total := x.tamanos(lista)
	x.limpiarTrabajo(lista)
	if falta := x.falta(carpeta, total); falta > 0 {
		m := mensajeEspacio(carpeta, falta)
		if x.reinstalar {
			m = msgEspacioVieja + " " + m
		}
		return x.fallo(m, fmt.Errorf("faltan %d bytes", falta))
	}

	ova := filepath.Join(x.dep.Trabajo, lista.Completa.Nombre)
	if x.ovaLista(lista, ova) {
		x.log("la OVA ya estaba unida y comprobada: %s", ova)
	} else if err := x.descargarYUnir(lista, tams, total, ova); err != nil {
		return err
	}

	if x.ctx.Err() != nil {
		return x.detenido()
	}
	if x.reinstalar {
		if err := x.borrarAnterior(); err != nil {
			return err
		}
	}
	if err := x.importar(ova, carpeta); err != nil {
		return err
	}
	x.intercambio()
	x.acceso()
	if err := os.RemoveAll(x.dep.Trabajo); err != nil {
		x.log("no se pudo borrar la carpeta de descarga: %v", err)
	}
	x.cerrarBarra()
	if err := vbox.Arrancar(x.e, Nombre); err != nil {
		return x.fallo(msgNoArranca, err)
	}
	x.log("instalación terminada")
	return nil
}

// ovaLista dice si la OVA unida de una ejecución anterior sirve. Solo se
// renombra a ova tras comprobarla, y unir borra las partes después; si aún
// queda alguna parte, la OVA no es de fiar y se borra.
func (x *ejecucion) ovaLista(l sumas.Lista, ova string) bool {
	if _, err := os.Stat(ova); err != nil {
		return false
	}
	for _, p := range l.Partes {
		if _, err := os.Stat(filepath.Join(x.dep.Trabajo, p.Nombre)); err == nil {
			x.log("hay una OVA y también partes: se vuelve a unir")
			os.Remove(ova)
			return false
		}
	}
	return true
}

// borrarAnterior borra la máquina que se reinstala, si sigue apagada.
func (x *ejecucion) borrarAnterior() error {
	existe, encendida, err := vbox.Estado(x.e, Nombre)
	if err != nil {
		return x.fallo(msgVBoxMudo, err)
	}
	if encendida {
		return x.fallo(msgAbierta, errors.New("la máquina anterior está encendida"))
	}
	if !existe {
		return nil
	}
	x.log("se borra la máquina anterior")
	if err := vbox.Borrar(x.e, Nombre); err != nil {
		return x.fallo(msgNoBorra, err)
	}
	return nil
}

func (x *ejecucion) leerSumas() (sumas.Lista, error) {
	req, err := http.NewRequestWithContext(x.ctx, http.MethodGet, x.dep.URLBase+"SHA256SUMS", nil)
	if err != nil {
		return sumas.Lista{}, err
	}
	resp, err := x.dep.Cliente.Do(req)
	if err != nil {
		return sumas.Lista{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return sumas.Lista{}, fmt.Errorf("SHA256SUMS: el servidor respondió %s", resp.Status)
	}
	return sumas.Leer(io.LimitReader(resp.Body, 1<<20))
}

// nombreValido exige un nombre de archivo simple, sin carpetas ni unidad.
func nombreValido(n string) bool {
	return n != "" && filepath.Base(n) == n && !strings.Contains(n, "..") && !strings.ContainsAny(n, `/\:`)
}

func validar(l sumas.Lista) error {
	if len(l.Partes) == 0 {
		return errors.New("SHA256SUMS no tiene partes")
	}
	for _, a := range append([]sumas.Archivo{l.Completa}, l.Partes...) {
		if !nombreValido(a.Nombre) {
			return fmt.Errorf("nombre no válido en SHA256SUMS: %q", a.Nombre)
		}
	}
	return nil
}

func (x *ejecucion) urlDe(nombre string) string { return x.dep.URLBase + url.PathEscape(nombre) }

// tamanos pide a cada parte su tamaño (Range: bytes=0-0, que también vale
// tras la redirección de GitHub). total es -1 si alguno no se sabe.
func (x *ejecucion) tamanos(l sumas.Lista) ([]int64, int64) {
	tams := make([]int64, len(l.Partes))
	var total int64
	for i, p := range l.Partes {
		n, err := x.tamano(x.urlDe(p.Nombre))
		if err != nil {
			x.log("no se sabe el tamaño de %s: %v", p.Nombre, err)
			n = -1
		}
		tams[i] = n
		if n < 0 || total < 0 {
			total = -1
		} else {
			total += n
		}
	}
	x.log("tamaño de la OVA: %d bytes", total)
	return tams, total
}

func (x *ejecucion) tamano(u string) (int64, error) {
	req, err := http.NewRequestWithContext(x.ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, err := x.dep.Cliente.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusPartialContent:
		cr := resp.Header.Get("Content-Range")
		if i := strings.LastIndex(cr, "/"); i >= 0 {
			return strconv.ParseInt(cr[i+1:], 10, 64)
		}
	case http.StatusOK:
		if resp.ContentLength >= 0 {
			return resp.ContentLength, nil
		}
	}
	return 0, fmt.Errorf("respuesta sin tamaño: %s %q", resp.Status, resp.Header.Get("Content-Range"))
}

// limpiarTrabajo borra de la carpeta de descarga lo que no es de esta release
// (restos de una versión anterior) y conserva lo que sirve para reanudar.
func (x *ejecucion) limpiarTrabajo(l sumas.Lista) {
	sirve := map[string]bool{l.Completa.Nombre: true}
	for _, p := range l.Partes {
		sirve[p.Nombre] = true
		sirve[p.Nombre+".descargando"] = true
	}
	entradas, _ := os.ReadDir(x.dep.Trabajo)
	for _, e := range entradas {
		if !sirve[e.Name()] {
			x.log("se borra un resto de otra descarga: %s", e.Name())
			os.RemoveAll(filepath.Join(x.dep.Trabajo, e.Name()))
		}
	}
}

// yaDescargado suma lo que ocupa la carpeta de descarga.
func (x *ejecucion) yaDescargado() int64 {
	var n int64
	entradas, _ := os.ReadDir(x.dep.Trabajo)
	for _, e := range entradas {
		if info, err := e.Info(); err == nil && !e.IsDir() {
			n += info.Size()
		}
	}
	return n
}

// falta devuelve los bytes que faltan en la unidad de ruta para terminar la
// instalación (0 o menos si caben). Si no se puede medir, no bloquea.
func (x *ejecucion) falta(ruta string, total int64) int64 {
	tam := total
	if tam <= 0 {
		tam = tamEstimado
	}
	necesario := equipo.EspacioNecesario(tam)
	if strings.EqualFold(filepath.VolumeName(ruta), filepath.VolumeName(x.dep.Trabajo)) {
		necesario -= x.yaDescargado()
	}
	libre, err := x.dep.EspacioLibre(existente(ruta))
	if err != nil {
		x.log("no se pudo medir el espacio libre en %s: %v", ruta, err)
		return 0
	}
	x.log("espacio en %s: libre %d, necesario %d", ruta, libre, necesario)
	return necesario - libre
}

// existente devuelve ruta o la primera carpeta madre que existe.
func existente(ruta string) string {
	for d := ruta; ; d = filepath.Dir(d) {
		if _, err := os.Stat(d); err == nil || filepath.Dir(d) == d {
			return d
		}
	}
}

func mensajeEspacio(ruta string, falta int64) string {
	disco := filepath.VolumeName(ruta)
	if disco != "" {
		disco = " " + disco
	}
	que := fmt.Sprintf("Faltan %d GB libres en el disco%s.", (falta+gib-1)/gib, disco)
	if falta <= 0 {
		que = fmt.Sprintf("El disco%s está lleno.", disco)
	}
	return que + " Libera espacio (por ejemplo, vacía la papelera) y vuelve a abrir el instalador: lo que ya se haya descargado no se pierde."
}

// descargarYUnir baja las partes y las une en ova. Si la OVA unida no cuadra
// con su suma, lo vuelve a descargar todo una vez.
func (x *ejecucion) descargarYUnir(l sumas.Lista, tams []int64, total int64, ova string) error {
	if err := os.MkdirAll(x.dep.Trabajo, 0o755); err != nil {
		return x.fallo(msgNoUne, err)
	}
	rutas := make([]string, len(l.Partes))
	for i, p := range l.Partes {
		rutas[i] = filepath.Join(x.dep.Trabajo, p.Nombre)
	}
	for vuelta := 1; ; vuelta++ {
		if err := x.descargar(l, tams, total); err != nil {
			return err
		}
		x.mostrar("Comprobando la descarga…", 1)
		// Se une con otro nombre: si el proceso muere a mitad, no queda una
		// OVA cortada que parezca buena (limpiarTrabajo borra el resto).
		err := unir.Partes(rutas, ova+".uniendo", l.Completa.SHA256)
		if err == nil {
			if err := os.Rename(ova+".uniendo", ova); err != nil {
				return x.fallo(msgNoUne, err)
			}
			return nil
		}
		switch {
		case errors.Is(err, unir.ErrSuma) && vuelta == 1:
			x.log("la OVA unida no cuadra; se descarga todo otra vez")
		case errors.Is(err, unir.ErrSuma):
			return x.fallo(msgDanada, err)
		case descarga.EsSinEspacio(err):
			return x.fallo(mensajeEspacio(x.dep.Trabajo, x.falta(x.dep.Trabajo, total)), err)
		default:
			return x.fallo(msgNoUne, err)
		}
	}
}

func (x *ejecucion) descargar(l sumas.Lista, tams []int64, total int64) error {
	var antes int64 // bytes de las partes ya terminadas
	x.ultimo = -1
	for i, p := range l.Partes {
		destino := filepath.Join(x.dep.Trabajo, p.Nombre)
		if st, err := os.Stat(destino); err == nil {
			// Bajar solo renombra a destino tras comprobar la suma.
			antes += st.Size()
			continue
		}
		progreso := func(hechos, tot int64) {
			switch {
			case total > 0:
				x.progreso(antes+hechos, total)
			case tot > 0:
				x.progreso(hechos, tot)
			}
		}
		var err error
		pausa := primeraPausa
		danadas := 0 // una parte que llega dañada se baja de nuevo solo una vez
		for intento := 1; intento <= intentosDescarga; intento++ {
			err = bajar(x.ctx, x.dep.Cliente, x.urlDe(p.Nombre), destino, p.SHA256, progreso)
			if errors.Is(err, descarga.ErrSuma) {
				danadas++
			}
			if err == nil || x.ctx.Err() != nil || descarga.EsSinEspacio(err) || danadas >= 2 {
				break
			}
			x.log("%s, intento %d de %d: %v", p.Nombre, intento, intentosDescarga, err)
			if intento < intentosDescarga {
				x.dep.Esperar(pausa)
				pausa *= 2
			}
		}
		switch {
		case err == nil:
		case x.ctx.Err() != nil:
			return x.detenido()
		case descarga.EsSinEspacio(err):
			return x.fallo(mensajeEspacio(x.dep.Trabajo, x.falta(x.dep.Trabajo, total)), err)
		case errors.Is(err, descarga.ErrSuma):
			return x.fallo(msgDanada, err)
		default:
			return x.fallo(msgCortada, err)
		}
		if st, err := os.Stat(destino); err == nil {
			antes += st.Size()
		} else if i < len(tams) && tams[i] > 0 {
			antes += tams[i]
		}
	}
	return nil
}

func (x *ejecucion) progreso(hechos, total int64) {
	f := float64(hechos) / float64(total)
	if m := int(f * 1000); m != x.ultimo {
		x.ultimo = m
		x.mostrar(fmt.Sprintf("Descargando Mint Docente: %s de %s", gb(hechos), gb(total)), f)
	}
}

func gb(n int64) string {
	return strings.Replace(fmt.Sprintf("%.1f GB", float64(n)/1e9), ".", ",", 1)
}

func (x *ejecucion) importar(ova, carpeta string) error {
	memTotal, err := x.dep.MemoriaTotalMB()
	if err != nil {
		x.log("no se pudo medir la memoria: %v; se supone 8 GB", err)
		memTotal = 8192
	}
	mem, cpus, lenta := equipo.Ajuste(memTotal, x.dep.Nucleos)
	x.log("memoria del equipo %d MB, %d núcleos: máquina con %d MB y %d CPU", memTotal, x.dep.Nucleos, mem, cpus)
	if lenta {
		x.aviso(msgPocaMemoria)
	}
	x.mostrar("Instalando Mint Docente en VirtualBox. Puede tardar unos minutos…", 1)
	if err := vbox.Importar(x.e, ova, Nombre, carpeta, mem, cpus); err != nil {
		if e2 := vbox.Borrar(x.e, Nombre); e2 != nil {
			x.log("no se pudo borrar la máquina a medias: %v", e2)
		}
		return x.fallo(msgNoImporta, err)
	}
	return nil
}

func (x *ejecucion) intercambio() {
	docs, err := x.dep.Documentos()
	dir := filepath.Join(docs, "Intercambio")
	if err == nil {
		err = os.MkdirAll(dir, 0o755)
	}
	if err == nil {
		err = vbox.Compartir(x.e, Nombre, dir)
	}
	if err != nil {
		x.log("Intercambio (%s): %v", dir, err)
		x.aviso(msgSinInterc)
	}
}

func (x *ejecucion) acceso() {
	destino := filepath.Join(filepath.Dir(x.rutaVBox), "VirtualBoxVM.exe")
	if err := x.dep.AccesoDirecto(destino, fmt.Sprintf("--startvm %q", Nombre)); err != nil {
		x.log("acceso directo: %v", err)
		x.aviso(msgSinAcceso)
	}
}

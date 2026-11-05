// Package vbox da órdenes a VirtualBox (VBoxManage): localiza la instalación,
// lee su versión e importa, comparte, arranca y borra la máquina. Cada ruta
// viaja como un único argumento, sin pasar por ningún intérprete de órdenes.
package vbox

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrNoInstalado indica que este equipo no tiene VirtualBox.
var ErrNoInstalado = errors.New("VirtualBox no está instalado")

// Ejecutor lanza VBoxManage con los argumentos dados y devuelve su salida.
type Ejecutor interface {
	Ejecutar(args ...string) (salida string, err error)
}

// Real ejecuta el VBoxManage que está en Ruta, sin mostrar consola.
type Real struct{ Ruta string }

// Ejecutar lanza VBoxManage; si falla, el error incluye la salida.
func (r Real) Ejecutar(args ...string) (string, error) {
	cmd := exec.Command(r.Ruta, args...)
	ocultarVentana(cmd)
	sal, err := cmd.CombinedOutput()
	if err != nil {
		return string(sal), fmt.Errorf("VBoxManage %s: %w: %s", firstArg(args), err, strings.TrimSpace(string(sal)))
	}
	return string(sal), nil
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// Version es la versión de VirtualBox (mayor.menor.parche).
type Version struct{ Mayor, Secundaria, Parche int }

// Menor dice si v es anterior a otra.
func (v Version) Menor(otra Version) bool {
	if v.Mayor != otra.Mayor {
		return v.Mayor < otra.Mayor
	}
	if v.Secundaria != otra.Secundaria {
		return v.Secundaria < otra.Secundaria
	}
	return v.Parche < otra.Parche
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Mayor, v.Secundaria, v.Parche)
}

// LeerVersion interpreta la salida de «VBoxManage --version», p. ej.
// 7.2.4r170995, tolerando sufijos como _Ubuntu o _BETA. (Se llama así y no
// Version porque ese nombre es el del tipo.)
func LeerVersion(e Ejecutor) (Version, error) {
	sal, err := e.Ejecutar("--version")
	if err != nil {
		return Version{}, err
	}
	return interpretarVersion(sal)
}

func interpretarVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	fin := 0
	for fin < len(s) && (s[fin] == '.' || (s[fin] >= '0' && s[fin] <= '9')) {
		fin++
	}
	partes := strings.Split(s[:fin], ".")
	var n [3]int
	for i := 0; i < len(partes) && i < 3; i++ {
		x, err := strconv.Atoi(partes[i])
		if err != nil {
			return Version{}, fmt.Errorf("versión de VirtualBox ilegible: %q", s)
		}
		n[i] = x
	}
	return Version{n[0], n[1], n[2]}, nil
}

// Estado dice si la máquina existe y si está encendida.
func Estado(e Ejecutor, nombre string) (existe, encendida bool, err error) {
	sal, err := e.Ejecutar("list", "vms")
	if err != nil {
		return false, false, err
	}
	existe = contiene(sal, nombre)
	sal, err = e.Ejecutar("list", "runningvms")
	if err != nil {
		return false, false, err
	}
	return existe, contiene(sal, nombre), nil
}

// contiene busca una línea «"nombre" {uuid}».
func contiene(salida, nombre string) bool {
	for _, l := range strings.Split(salida, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), `"`+nombre+`" {`) {
			return true
		}
	}
	return false
}

// Hooks que sustituyen las pruebas (y, en Windows, los códigos de página).
var (
	// decodificadores convierten la salida cruda de VBoxManage, en la página
	// ANSI y después en la OEM, a UTF-8. Vacío fuera de Windows.
	decodificadores []func(string) string
	existeRuta      = func(r string) bool { _, err := os.Stat(r); return err == nil }
	perfilUsuario   = func() string { return os.Getenv("USERPROFILE") }
)

// CarpetaMaquinas devuelve la carpeta de máquinas por defecto de VirtualBox.
// VBoxManage escribe por la tubería en la página de códigos de Windows, no en
// UTF-8: se prueba UTF-8, luego ANSI y luego OEM, y se queda la primera
// lectura cuya carpeta (o su carpeta madre) existe. Si ninguna, usa
// %USERPROFILE%\VirtualBox VMs.
func CarpetaMaquinas(e Ejecutor) (string, error) {
	sal, err := e.Ejecutar("list", "systemproperties")
	if err != nil {
		return "", err
	}
	if _, ok := carpetaDeSalida(sal); !ok {
		return "", errors.New("VirtualBox no indica su carpeta de máquinas")
	}
	return elegirCarpeta(sal, decodificadores, existeRuta, perfilUsuario())
}

func carpetaDeSalida(sal string) (string, bool) {
	const clave = "Default machine folder:"
	for _, l := range strings.Split(sal, "\n") {
		if r, ok := strings.CutPrefix(l, clave); ok {
			if c := strings.TrimSpace(r); c != "" {
				return c, true
			}
		}
	}
	return "", false
}

func elegirCarpeta(sal string, decs []func(string) string, existe func(string) bool, perfil string) (string, error) {
	// Las lecturas se calculan por orden y de forma perezosa.
	lecturas := []func() string{}
	if utf8.ValidString(sal) {
		lecturas = append(lecturas, func() string { return sal })
	}
	for _, d := range decs {
		lecturas = append(lecturas, func() string { return d(sal) })
	}
	for _, leer := range lecturas {
		c, ok := carpetaDeSalida(leer())
		if !ok || !utf8.ValidString(c) {
			continue
		}
		if existe(c) || existe(madre(c)) {
			return c, nil
		}
	}
	if perfil == "" {
		return "", errors.New("no se encuentra la carpeta de máquinas de VirtualBox")
	}
	return perfil + `\VirtualBox VMs`, nil
}

// madre devuelve la carpeta que contiene a c (separador \ o /).
func madre(c string) string {
	c = strings.TrimRight(c, `\/`)
	if i := strings.LastIndexAny(c, `\/`); i > 0 {
		return c[:i]
	}
	return c
}

// Importar importa la OVA como una máquina nueva.
func Importar(e Ejecutor, ova, nombre, carpeta string, memoriaMB, cpus int) error {
	_, err := e.Ejecutar("import", ova, "--vsys", "0", "--vmname", nombre,
		"--basefolder", carpeta, "--memory", strconv.Itoa(memoriaMB), "--cpus", strconv.Itoa(cpus))
	return err
}

// Compartir añade la carpeta compartida Intercambio con montaje automático.
func Compartir(e Ejecutor, nombre, ruta string) error {
	_, err := e.Ejecutar("sharedfolder", "add", nombre, "--name", "Intercambio", "--hostpath", ruta, "--automount")
	return err
}

// Arrancar enciende la máquina con ventana.
func Arrancar(e Ejecutor, nombre string) error {
	_, err := e.Ejecutar("startvm", nombre, "--type", "gui")
	return err
}

// Borrar da de baja la máquina y borra sus archivos.
func Borrar(e Ejecutor, nombre string) error {
	_, err := e.Ejecutar("unregistervm", nombre, "--delete")
	return err
}

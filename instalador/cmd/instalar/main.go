// InstalarMintDocente: instalador para Windows (subsistema GUI, sin consola).
// Descarga la máquina Mint Docente, la importa en VirtualBox y la arranca.
// Todo mensaje va a una ventana y a %LOCALAPPDATA%\MintDocente\instalador.log;
// nunca a la salida estándar.
package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/balejosg/mint-docente/instalador/internal/acceso"
	"github.com/balejosg/mint-docente/instalador/internal/descarga"
	"github.com/balejosg/mint-docente/instalador/internal/equipo"
	"github.com/balejosg/mint-docente/instalador/internal/flujo"
	"github.com/balejosg/mint-docente/instalador/internal/registro"
	"github.com/balejosg/mint-docente/instalador/internal/ui"
	"github.com/balejosg/mint-docente/instalador/internal/vbox"
)

const origen = "https://github.com/balejosg/mint-docente/releases/latest/download/"

func main() { os.Exit(ejecutar()) }

// carpetaBase es %LOCALAPPDATA%\MintDocente.
func carpetaBase() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		local, _ = os.UserCacheDir()
	}
	if local == "" {
		local = os.TempDir()
	}
	return filepath.Join(local, "MintDocente")
}

func ejecutar() (codigo int) {
	base := carpetaBase()
	rutaLog := filepath.Join(base, "instalador.log")
	lg, cierre := registro.Abrir(rutaLog)
	defer cierre.Close()
	log := registro.Func(lg)

	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()
	ventanas := ui.Zenity{Detener: cancelar, Log: log}

	defer func() {
		if r := recover(); r != nil {
			lg.Printf("fallo inesperado: %v\n%s", r, debug.Stack())
			ventanas.Error(flujo.ConAyuda("Algo ha fallado en el instalador. Vuelve a abrirlo.", rutaLog))
			codigo = 2
		}
	}()

	dep := flujo.Dependencias{
		UI:             ventanas,
		Localizar:      vbox.Localizar,
		NuevoEjecutor:  func(ruta string) vbox.Ejecutor { return vbox.Real{Ruta: ruta} },
		Cliente:        descarga.Cliente(),
		URLBase:        origen,
		Trabajo:        filepath.Join(base, "descarga"),
		Registro:       rutaLog,
		Log:            log,
		MemoriaTotalMB: equipo.MemoriaTotalMB,
		Nucleos:        runtime.NumCPU(),
		EspacioLibre:   equipo.EspacioLibre,
		Documentos:     equipo.Documentos,
		AccesoDirecto:  acceso.Crear,
		Esperar:        time.Sleep,
	}
	log("Windows %s/%s, %d núcleos", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
	if err := flujo.Ejecutar(ctx, dep); err != nil {
		log("terminado sin instalar: %v", err)
		return 1
	}
	return 0
}

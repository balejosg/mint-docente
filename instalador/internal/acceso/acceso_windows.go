package acceso

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

// Crear deja en el escritorio del usuario (respetando redirecciones, p. ej.
// a OneDrive) un acceso directo «Mint Docente.lnk» que abre destino con
// argumentos, usando WScript.Shell.CreateShortcut. Si ya existe, lo rehace.
func Crear(destino, argumentos string) (err error) {
	escritorio, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, 0)
	if err != nil {
		return fmt.Errorf("carpeta del escritorio: %w", err)
	}
	lnk := filepath.Join(escritorio, Nombre+".lnk")

	// COM exige usar el objeto en el mismo hilo en que se inicializa.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		var oe *ole.OleError
		switch {
		case errors.As(err, &oe) && oe.Code() == 1: // S_FALSE: ya estaba inicializado
			defer ole.CoUninitialize()
		case errors.As(err, &oe) && oe.Code() == 0x80010106: // RPC_E_CHANGED_MODE: sirve igual
		default:
			return fmt.Errorf("CoInitializeEx: %w", err)
		}
	} else {
		defer ole.CoUninitialize()
	}

	unk, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return fmt.Errorf("WScript.Shell: %w", err)
	}
	defer unk.Release()
	shell, err := unk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("WScript.Shell: %w", err)
	}
	defer shell.Release()

	v, err := oleutil.CallMethod(shell, "CreateShortcut", lnk)
	if err != nil {
		return fmt.Errorf("CreateShortcut %s: %w", lnk, err)
	}
	defer v.Clear()
	sc := v.ToIDispatch()

	for _, p := range []struct{ nombre, valor string }{
		{"TargetPath", destino},
		{"Arguments", argumentos},
		{"WorkingDirectory", filepath.Dir(destino)},
		{"IconLocation", destino + ",0"},
		{"Description", "Abrir Mint Docente"},
	} {
		if _, err := oleutil.PutProperty(sc, p.nombre, p.valor); err != nil {
			return fmt.Errorf("acceso directo, %s: %w", p.nombre, err)
		}
	}
	if _, err := oleutil.CallMethod(sc, "Save"); err != nil {
		return fmt.Errorf("guardar %s: %w", lnk, err)
	}
	return nil
}

package equipo

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type memoriaEstado struct {
	longitud        uint32
	carga           uint32
	totalFisica     uint64
	disponibleFis   uint64
	totalPagina     uint64
	disponiblePag   uint64
	totalVirtual    uint64
	disponibleVirt  uint64
	disponibleExtra uint64
}

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	globalMemoryStat = kernel32.NewProc("GlobalMemoryStatusEx")
)

// MemoriaTotalMB devuelve la memoria física total en MB (GlobalMemoryStatusEx).
func MemoriaTotalMB() (int, error) {
	var m memoriaEstado
	m.longitud = uint32(unsafe.Sizeof(m))
	r, _, err := globalMemoryStat.Call(uintptr(unsafe.Pointer(&m)))
	if r == 0 {
		return 0, err
	}
	return int(m.totalFisica / (1 << 20)), nil
}

// EspacioLibre devuelve los bytes libres para el usuario en la unidad de ruta
// (GetDiskFreeSpaceEx).
func EspacioLibre(ruta string) (int64, error) {
	p, err := windows.UTF16PtrFromString(ruta)
	if err != nil {
		return 0, err
	}
	var libre, total, libreTotal uint64
	if err := windows.GetDiskFreeSpaceEx(p, &libre, &total, &libreTotal); err != nil {
		return 0, err
	}
	return int64(libre), nil
}

// Documentos devuelve la carpeta Documentos del usuario, respetando las
// redirecciones (carpeta conocida FOLDERID_Documents).
func Documentos() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_Documents, 0)
}

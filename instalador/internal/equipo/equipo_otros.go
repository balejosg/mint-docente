//go:build !windows

package equipo

import "errors"

var errSoloWindows = errors.New("solo disponible en Windows")

// MemoriaTotalMB solo funciona en Windows; aquí existe para que compile.
func MemoriaTotalMB() (int, error) { return 0, errSoloWindows }

// EspacioLibre solo funciona en Windows; aquí existe para que compile.
func EspacioLibre(ruta string) (int64, error) { return 0, errSoloWindows }

// Documentos solo funciona en Windows; aquí existe para que compile.
func Documentos() (string, error) { return "", errSoloWindows }

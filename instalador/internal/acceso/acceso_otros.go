//go:build !windows

package acceso

import "errors"

// Crear solo funciona en Windows; aquí existe para que compile.
func Crear(destino, argumentos string) error {
	return errors.New("el acceso directo solo se crea en Windows")
}

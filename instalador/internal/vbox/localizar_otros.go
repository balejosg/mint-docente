//go:build !windows

package vbox

// Localizar solo funciona en Windows; aquí existe para que compilen las pruebas.
func Localizar() (string, error) { return "", ErrNoInstalado }

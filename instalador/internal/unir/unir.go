// Package unir junta las partes de la OVA en un solo archivo, sin cargarlas
// en memoria, y comprueba el sha256 del resultado.
package unir

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrSuma indica que el archivo completo no coincide con el sha256 esperado.
var ErrSuma = errors.New("la suma de comprobación no coincide")

// Partes concatena las partes en orden en destino y comprueba la suma de la
// completa. Las partes solo se borran cuando la suma cuadra. Si no cuadra
// devuelve ErrSuma y deja sin destino ni partes, para que el flujo vuelva a
// descargar. Ante cualquier otro error (disco lleno, parte ausente, fallo de
// lectura) borra solo el destino a medias y conserva todas las partes.
func Partes(partes []string, destino string, sumaEsperada string) error {
	out, err := os.Create(destino)
	if err != nil {
		return err
	}
	h := sha256.New()
	fallo := func(e error) error {
		out.Close()
		os.Remove(destino)
		return e
	}
	for _, p := range partes {
		if err := anadir(out, h, p); err != nil {
			return fallo(err)
		}
	}
	if err := out.Close(); err != nil {
		os.Remove(destino)
		return fmt.Errorf("al cerrar %s: %w", destino, err)
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), sumaEsperada) {
		os.Remove(destino)
		for _, p := range partes {
			os.Remove(p)
		}
		return ErrSuma
	}
	for _, p := range partes {
		os.Remove(p) // si no se puede ahora, no invalida la unión ya verificada
	}
	return nil
}

func anadir(out io.Writer, h io.Writer, ruta string) error {
	f, err := os.Open(ruta)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(io.MultiWriter(out, h), f); err != nil {
		return fmt.Errorf("al unir %s: %w", ruta, err)
	}
	return nil
}

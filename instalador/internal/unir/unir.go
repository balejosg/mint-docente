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

// Partes concatena las partes en orden en destino, borrando cada una después
// de añadirla. Si la suma de la completa no cuadra devuelve ErrSuma y deja sin
// destino ni partes, para que el flujo vuelva a descargar.
func Partes(partes []string, destino string, sha256Esperado string) (err error) {
	out, err := os.Create(destino)
	if err != nil {
		return err
	}
	h := sha256.New()
	defer func() {
		if err != nil {
			out.Close()
			os.Remove(destino)
			for _, p := range partes {
				os.Remove(p)
			}
		}
	}()
	for _, p := range partes {
		if err = anadir(out, h, p); err != nil {
			return err
		}
		if err = os.Remove(p); err != nil {
			return err
		}
	}
	if err = out.Close(); err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), sha256Esperado) {
		return ErrSuma
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

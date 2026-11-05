// Package registro escribe instalador.log, donde va el detalle técnico que
// el docente puede enviar si necesita ayuda.
package registro

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Abrir abre (o crea) el registro en ruta y añade al final. Si no se puede,
// devuelve un registro que no escribe nada, para que el instalador siga.
func Abrir(ruta string) (*log.Logger, io.Closer) {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err == nil {
		if f, err := os.OpenFile(ruta, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
			l := log.New(f, "", log.LstdFlags)
			l.Println("----- instalador de Mint Docente -----")
			return l, f
		}
	}
	return log.New(io.Discard, "", 0), nada{}
}

type nada struct{}

func (nada) Close() error { return nil }

// Func adapta el registro a la firma que usa flujo.
func Func(l *log.Logger) func(string, ...any) {
	return func(f string, a ...any) { l.Output(2, fmt.Sprintf(f, a...)) }
}

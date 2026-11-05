// Package ui muestra las ventanas del instalador con zenity: avisos, errores,
// preguntas y la barra de progreso. Nada se escribe en una consola.
package ui

import (
	"errors"
	"sync"

	"github.com/ncruces/zenity"

	"github.com/balejosg/mint-docente/instalador/internal/flujo"
)

const titulo = "Instalar Mint Docente"

// Zenity implementa flujo.UI. Si el docente cierra la barra de progreso, se
// llama a Detener para parar la instalación.
type Zenity struct {
	Detener func()
	// Log recibe los fallos de las propias ventanas.
	Log func(formato string, args ...any)
}

func (z Zenity) log(f string, a ...any) {
	if z.Log != nil {
		z.Log(f, a...)
	}
}

// Aviso muestra un mensaje informativo y espera a que se cierre.
func (z Zenity) Aviso(texto string) {
	if err := zenity.Info(texto, zenity.Title(titulo), zenity.InfoIcon); err != nil && !errors.Is(err, zenity.ErrCanceled) {
		z.log("ventana de aviso: %v", err)
	}
}

// Error muestra un error y espera a que se cierre.
func (z Zenity) Error(texto string) {
	if err := zenity.Error(texto, zenity.Title(titulo), zenity.ErrorIcon); err != nil && !errors.Is(err, zenity.ErrCanceled) {
		z.log("ventana de error: %v", err)
	}
}

// Pregunta devuelve true si se pulsa si; cerrar la ventana cuenta como no.
func (z Zenity) Pregunta(texto, si, no string) bool {
	err := zenity.Question(texto, zenity.Title(titulo), zenity.QuestionIcon,
		zenity.OKLabel(si), zenity.CancelLabel(no))
	if err != nil && !errors.Is(err, zenity.ErrCanceled) {
		z.log("ventana de pregunta: %v", err)
	}
	return err == nil
}

const maximo = 1000

// Progreso abre una barra de progreso.
func (z Zenity) Progreso(t string) flujo.Barra {
	dlg, err := zenity.Progress(zenity.Title(t), zenity.MaxValue(maximo), zenity.CancelLabel("Detener"))
	if err != nil {
		z.log("barra de progreso: %v", err)
		return barraNula{}
	}
	b := &barra{dlg: dlg, cerrada: make(chan struct{})}
	go func() {
		<-dlg.Done()
		select {
		case <-b.cerrada: // la ha cerrado el instalador
		default: // la ha cerrado el docente o ha pulsado «Detener»
			if z.Detener != nil {
				z.Detener()
			}
		}
	}()
	return b
}

type barra struct {
	dlg     zenity.ProgressDialog
	cerrada chan struct{}
	una     sync.Once
}

func (b *barra) Actualizar(texto string, fraccion float64) {
	b.dlg.Text(texto)
	b.dlg.Value(int(min(max(fraccion, 0), 1) * maximo))
}

func (b *barra) Cerrar() {
	b.una.Do(func() {
		close(b.cerrada)
		b.dlg.Close()
	})
}

type barraNula struct{}

func (barraNula) Actualizar(string, float64) {}
func (barraNula) Cerrar()                    {}

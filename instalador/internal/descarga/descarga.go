// Package descarga baja archivos grandes por HTTP(S): reanuda lo que quedó a
// medias, usa el proxy del sistema y comprueba el sha256 al terminar.
package descarga

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-ieproxy"
)

var (
	// ErrSuma indica que lo descargado no coincide con el sha256 esperado.
	ErrSuma = errors.New("la suma de comprobación no coincide")
	// ErrSinEspacio indica que el disco se llenó al escribir.
	ErrSinEspacio = errors.New("no queda espacio en el disco")
)

const reintentos = 3

var esperaReintento = 2 * time.Second

// abrirParcial abre el archivo parcial; si anadir es falso lo vacía.
var abrirParcial = abrirArchivo

func abrirArchivo(ruta string, anadir bool) (io.WriteCloser, error) {
	flags := os.O_CREATE | os.O_WRONLY
	if anadir {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(ruta, flags, 0o644)
}

// Cliente devuelve un cliente HTTP con el proxy del sistema (también PAC en
// Windows) y hasta 3 reintentos con espera creciente ante errores de red.
func Cliente() *http.Client {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = ieproxy.GetProxyFunc()
	base.ResponseHeaderTimeout = 60 * time.Second
	return &http.Client{Transport: &reintentador{base: base}}
}

type reintentador struct{ base http.RoundTripper }

func (r *reintentador) RoundTrip(req *http.Request) (*http.Response, error) {
	espera := esperaReintento
	for i := 0; ; i++ {
		resp, err := r.base.RoundTrip(req)
		if err == nil || i >= reintentos-1 || req.Context().Err() != nil || (req.Body != nil && req.GetBody == nil) {
			return resp, err
		}
		select {
		case <-time.After(espera):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
		espera *= 2
	}
}

func sinEspacio(err error) bool {
	if errors.Is(err, syscall.ENOSPC) {
		return true
	}
	var en syscall.Errno
	// Windows: ERROR_HANDLE_DISK_FULL (39) y ERROR_DISK_FULL (112).
	return runtime.GOOS == "windows" && errors.As(err, &en) && (en == 39 || en == 112)
}

// Bajar descarga url a destino+".descargando", continuando con Range si ya
// existe, y al terminar comprueba el sha256 y renombra a destino.
func Bajar(ctx context.Context, cli *http.Client, url, destino, sha string, progreso func(hechos, total int64)) error {
	parcial := destino + ".descargando"
	var previo int64
	if st, err := os.Stat(parcial); err == nil {
		previo = st.Size()
	}
	if previo > 0 {
		_, err := intentar(ctx, cli, url, parcial, previo, progreso)
		if err == nil {
			return cerrar(parcial, destino, sha)
		}
		if !errors.Is(err, errReiniciar) {
			return err
		}
	}
	if err := bajarDesdeCero(ctx, cli, url, parcial, progreso); err != nil {
		return err
	}
	return cerrar(parcial, destino, sha)
}

var errReiniciar = errors.New("reiniciar")

func bajarDesdeCero(ctx context.Context, cli *http.Client, url, parcial string, progreso func(int64, int64)) error {
	_, err := intentar(ctx, cli, url, parcial, 0, progreso)
	return err
}

func intentar(ctx context.Context, cli *http.Client, url, parcial string, previo int64, progreso func(int64, int64)) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if previo > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", previo))
	}
	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	anadir := false
	total := resp.ContentLength
	switch {
	case previo > 0 && resp.StatusCode == http.StatusPartialContent:
		anadir = true
		if total >= 0 {
			total += previo
		}
	case resp.StatusCode == http.StatusOK:
		previo = 0
	case previo > 0 && resp.StatusCode == http.StatusRequestedRangeNotSatisfiable:
		return 0, errReiniciar
	default:
		return 0, fmt.Errorf("el servidor respondió %s", resp.Status)
	}
	w, err := abrirParcial(parcial, anadir)
	if err != nil {
		return 0, traducir(err)
	}
	hechos := previo
	if progreso != nil {
		progreso(hechos, total)
	}
	buf := make([]byte, 256*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				w.Close()
				return hechos, traducir(werr)
			}
			hechos += int64(n)
			if progreso != nil {
				progreso(hechos, total)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			w.Close()
			return hechos, rerr
		}
	}
	if err := w.Close(); err != nil {
		return hechos, traducir(err)
	}
	if total >= 0 && hechos != total {
		return hechos, io.ErrUnexpectedEOF
	}
	return hechos, nil
}

func traducir(err error) error {
	if sinEspacio(err) {
		return fmt.Errorf("%w: %v", ErrSinEspacio, err)
	}
	return err
}

func cerrar(parcial, destino, sha string) error {
	f, err := os.Open(parcial)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	f.Close()
	if err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), sha) {
		os.Remove(parcial)
		os.Remove(destino)
		return ErrSuma
	}
	return os.Rename(parcial, destino)
}

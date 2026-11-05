package descarga

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func datos() []byte { return bytes.Repeat([]byte("mint-docente-"), 5000) }

func suma(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// servidor sirve b; si cortar>0, la primera petición envía solo esa cantidad de
// bytes y corta la conexión. Registra la cabecera Range de cada petición.
func servidor(t *testing.T, b []byte, cortar int, conRange bool) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	rangos := &[]string{}
	primera := true
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*rangos = append(*rangos, r.Header.Get("Range"))
		corta := primera && cortar > 0
		primera = false
		mu.Unlock()
		inicio := 0
		if rg := r.Header.Get("Range"); rg != "" && conRange {
			fmt.Sscanf(rg, "bytes=%d-", &inicio)
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", inicio, len(b)-1, len(b)))
			w.Header().Set("Content-Length", strconv.Itoa(len(b)-inicio))
			w.WriteHeader(http.StatusPartialContent)
		} else {
			w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		}
		if corta {
			w.Write(b[:cortar])
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		}
		w.Write(b[inicio:])
	}))
	t.Cleanup(s.Close)
	return s, rangos
}

func TestReanuda(t *testing.T) {
	b := datos()
	s, rangos := servidor(t, b, 1000, true)
	dest := filepath.Join(t.TempDir(), "a.part1")
	if err := Bajar(context.Background(), s.Client(), s.URL, dest, suma(b), nil); err == nil {
		t.Fatal("la primera llamada debía fallar por el corte")
	}
	parcial, err := os.Stat(dest + ".descargando")
	if err != nil || parcial.Size() != 1000 {
		t.Fatalf("parcial: %v %v", parcial, err)
	}
	var ultimo int64
	if err := Bajar(context.Background(), s.Client(), s.URL, dest, suma(b), func(h, _ int64) { ultimo = h }); err != nil {
		t.Fatal(err)
	}
	if (*rangos)[1] != "bytes=1000-" {
		t.Fatalf("Range = %q", (*rangos)[1])
	}
	got, _ := os.ReadFile(dest)
	if !bytes.Equal(got, b) {
		t.Fatal("contenido distinto")
	}
	if ultimo != int64(len(b)) {
		t.Fatalf("progreso final %d", ultimo)
	}
	if _, err := os.Stat(dest + ".descargando"); !os.IsNotExist(err) {
		t.Fatal("queda .descargando")
	}
}

func TestServidorSinRange(t *testing.T) {
	b := datos()
	s, _ := servidor(t, b, 1000, false)
	dest := filepath.Join(t.TempDir(), "a.part1")
	Bajar(context.Background(), s.Client(), s.URL, dest, suma(b), nil)
	if err := Bajar(context.Background(), s.Client(), s.URL, dest, suma(b), nil); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dest)
	if !bytes.Equal(got, b) {
		t.Fatal("archivo corrupto")
	}
}

func TestSumaIncorrecta(t *testing.T) {
	b := datos()
	s, _ := servidor(t, b, 0, true)
	dest := filepath.Join(t.TempDir(), "a.part1")
	err := Bajar(context.Background(), s.Client(), s.URL, dest, suma([]byte("otra")), nil)
	if !errors.Is(err, ErrSuma) {
		t.Fatalf("err = %v", err)
	}
	for _, p := range []string{dest, dest + ".descargando"} {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			t.Fatalf("queda %s", p)
		}
	}
}

func TestRedireccion(t *testing.T) {
	b := datos()
	final, _ := servidor(t, b, 0, true)
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+"/objeto", http.StatusFound)
	}))
	defer redir.Close()
	dest := filepath.Join(t.TempDir(), "a.part1")
	if err := Bajar(context.Background(), redir.Client(), redir.URL, dest, suma(b), nil); err != nil {
		t.Fatal(err)
	}
}

type escritorLleno struct{ n int }

func (e *escritorLleno) Write(p []byte) (int, error) {
	return 0, &os.PathError{Op: "write", Path: "x", Err: syscall.ENOSPC}
}
func (e *escritorLleno) Close() error { return nil }

func TestDiscoLleno(t *testing.T) {
	b := datos()
	s, _ := servidor(t, b, 0, true)
	abrirParcial = func(string, bool) (io.WriteCloser, error) { return &escritorLleno{}, nil }
	defer func() { abrirParcial = abrirArchivo }()
	err := Bajar(context.Background(), s.Client(), s.URL, filepath.Join(t.TempDir(), "a"), suma(b), nil)
	if !errors.Is(err, ErrSinEspacio) {
		t.Fatalf("err = %v", err)
	}
}

func TestUsaProxy(t *testing.T) {
	b := datos()
	var visto string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visto = r.Host + r.URL.Path
		w.Write(b)
	}))
	defer proxy.Close()
	// Los hosts locales nunca pasan por proxy, así que se usa un nombre ficticio.
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	dest := filepath.Join(t.TempDir(), "a.part1")
	if err := Bajar(context.Background(), Cliente(), "http://servidor.ficticio.test/x/a.part1", dest, suma(b), nil); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(visto, "servidor.ficticio.test/x/") {
		t.Fatalf("el proxy vio %q", visto)
	}
}

func TestReintentaErroresDeRed(t *testing.T) {
	esperaReintento = time.Millisecond
	n := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n < 3 {
			panic(http.ErrAbortHandler)
		}
		w.Write([]byte("ok"))
	}))
	defer s.Close()
	c := Cliente()
	resp, err := c.Get(s.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if n != 3 {
		t.Fatalf("intentos %d", n)
	}
}

// Package sumas lee el SHA256SUMS de una release: una línea por parte
// «*.ova.partN» y una línea para la «.ova» completa.
package sumas

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Archivo struct {
	Nombre string
	SHA256 string
}

type Lista struct {
	Partes   []Archivo // ordenadas por el número N de .partN
	Completa Archivo
}

var (
	reLinea = regexp.MustCompile(`^([0-9a-fA-F]{64}) [ *](.+)$`)
	reParte = regexp.MustCompile(`\.ova\.part(\d+)$`)
)

func Leer(r io.Reader) (Lista, error) {
	var l Lista
	nums := map[string]int{}
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		txt := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(txt) == "" {
			continue
		}
		m := reLinea.FindStringSubmatch(txt)
		if m == nil {
			return Lista{}, fmt.Errorf("línea %d mal formada en SHA256SUMS", n)
		}
		a := Archivo{Nombre: m[2], SHA256: strings.ToLower(m[1])}
		switch p := reParte.FindStringSubmatch(a.Nombre); {
		case p != nil:
			num, err := strconv.Atoi(p[1])
			if err != nil {
				return Lista{}, fmt.Errorf("línea %d: número de parte no válido", n)
			}
			nums[a.Nombre] = num
			l.Partes = append(l.Partes, a)
		case strings.HasSuffix(a.Nombre, ".ova"):
			if l.Completa.Nombre != "" {
				return Lista{}, fmt.Errorf("SHA256SUMS tiene más de una .ova completa")
			}
			l.Completa = a
		default:
			return Lista{}, fmt.Errorf("línea %d: archivo inesperado %q", n, a.Nombre)
		}
	}
	if err := sc.Err(); err != nil {
		return Lista{}, err
	}
	if l.Completa.Nombre == "" {
		return Lista{}, fmt.Errorf("SHA256SUMS no incluye la .ova completa")
	}
	sort.Slice(l.Partes, func(i, j int) bool {
		return nums[l.Partes[i].Nombre] < nums[l.Partes[j].Nombre]
	})
	return l, nil
}

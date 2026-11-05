package vbox

import "golang.org/x/sys/windows"

func init() {
	decodificadores = []func(string) string{
		func(s string) string { return desdePagina(windows.GetACP(), s) },
		func(s string) string { return desdePagina(oem(), s) },
	}
}

func oem() uint32 {
	r, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetOEMCP").Call()
	return uint32(r)
}

// desdePagina convierte bytes de una página de códigos a UTF-8.
func desdePagina(pagina uint32, s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	n, err := windows.MultiByteToWideChar(pagina, 0, &b[0], int32(len(b)), nil, 0)
	if err != nil || n == 0 {
		return s
	}
	w := make([]uint16, n)
	if _, err := windows.MultiByteToWideChar(pagina, 0, &b[0], int32(len(b)), &w[0], n); err != nil {
		return s
	}
	return windows.UTF16ToString(w)
}

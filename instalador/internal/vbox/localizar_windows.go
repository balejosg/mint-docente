package vbox

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const exe = "VBoxManage.exe"

// Localizar busca VBoxManage.exe: registro (HKLM\SOFTWARE\Oracle\VirtualBox,
// InstallDir), después VBOX_MSI_INSTALL_PATH y después %ProgramFiles%.
func Localizar() (string, error) {
	var carpetas []string
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`,
		registry.QUERY_VALUE|registry.WOW64_64KEY); err == nil {
		if d, _, err := k.GetStringValue("InstallDir"); err == nil {
			carpetas = append(carpetas, d)
		}
		k.Close()
	}
	carpetas = append(carpetas, os.Getenv("VBOX_MSI_INSTALL_PATH"))
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		carpetas = append(carpetas, filepath.Join(pf, "Oracle", "VirtualBox"))
	}
	for _, d := range carpetas {
		if d == "" {
			continue
		}
		ruta := filepath.Join(d, exe)
		if st, err := os.Stat(ruta); err == nil && !st.IsDir() {
			return ruta, nil
		}
	}
	return "", ErrNoInstalado
}

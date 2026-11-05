package vbox

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// ocultarVentana evita que aparezca una consola al lanzar VBoxManage.
func ocultarVentana(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
}

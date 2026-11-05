package vbox

import (
	"os/exec"
	"syscall"
)

// ocultarVentana evita que aparezca una consola al lanzar VBoxManage.
func ocultarVentana(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

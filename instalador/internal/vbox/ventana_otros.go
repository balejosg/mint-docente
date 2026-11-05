//go:build !windows

package vbox

import "os/exec"

func ocultarVentana(*exec.Cmd) {}

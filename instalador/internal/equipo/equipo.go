// Package equipo mide los recursos del ordenador para dimensionar la máquina virtual.
package equipo

const (
	gib = int64(1) << 30

	memMinMB   = 2048
	memMaxMB   = 4096
	pasoMemMB  = 512
	maxCPUs    = 4
	umbralLent = 6144
)

// Ajuste calcula la memoria (MB) y los procesadores de la máquina virtual:
// la mitad de la memoria, en múltiplos de 512 y entre 2048 y 4096, y la mitad
// de los procesadores lógicos, entre 1 y 4. lenta indica menos de 6144 MB en total.
func Ajuste(memoriaTotalMB int, nucleos int) (memoriaMB, cpus int, lenta bool) {
	memoriaMB = memoriaTotalMB / 2 / pasoMemMB * pasoMemMB
	memoriaMB = min(max(memoriaMB, memMinMB), memMaxMB)
	cpus = min(max(nucleos/2, 1), maxCPUs)
	return memoriaMB, cpus, memoriaTotalMB < umbralLent
}

// EspacioNecesario es el espacio libre que hace falta para una OVA de tamOVA
// bytes: la OVA, su copia desempaquetada y 12 GiB para el disco de la máquina.
func EspacioNecesario(tamOVA int64) int64 {
	return 2*tamOVA + 12*gib
}

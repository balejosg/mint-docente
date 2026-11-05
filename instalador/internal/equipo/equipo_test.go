package equipo

import "testing"

func TestAjuste(t *testing.T) {
	casos := []struct {
		mem, nuc int
		memMB    int
		cpus     int
		lenta    bool
	}{
		{8192, 4, 4096, 2, false},
		{4096, 2, 2048, 1, true},
		{16384, 16, 4096, 4, false},
		{6000, 2, 2560, 1, true},
		{6144, 1, 3072, 1, false},
	}
	for _, c := range casos {
		m, p, l := Ajuste(c.mem, c.nuc)
		if m != c.memMB || p != c.cpus || l != c.lenta {
			t.Errorf("Ajuste(%d, %d) = (%d, %d, %v), quiero (%d, %d, %v)",
				c.mem, c.nuc, m, p, l, c.memMB, c.cpus, c.lenta)
		}
	}
}

func TestEspacioNecesario(t *testing.T) {
	const gb = 1000 * 1000 * 1000
	const gib = 1 << 30
	got := EspacioNecesario(3500 * 1000 * 1000)
	if want := int64(7*gb + 12*gib); got != want {
		t.Errorf("EspacioNecesario = %d, quiero %d", got, want)
	}
}

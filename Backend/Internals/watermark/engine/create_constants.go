package watermark

import (
	"math"
)

type Constants struct {
	U            int
	V            int
	Const_matrix [8][8]float64
	Uk           float64
	Vk           float64
	Nc           float64
}

func (d *Constants) FindValueOptimized(block [][]float64) float64 {

	// Calculate DCT coefficient
	total := 0.0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			total += block[x][y] * d.Const_matrix[y][x]
		}
	}

	return d.Nc * total
}

func CreateConstant(u int, v int) *Constants {
	d := new(Constants)

	d.U = u
	d.V = v
	// Pre-calculate u and v normalization coefficients
	if u == 0 {
		d.Uk = math.Sqrt(1.0 / 8.0)
	} else {
		d.Uk = math.Sqrt(2.0 / 8.0)
	}

	if v == 0 {
		d.Vk = math.Sqrt(1.0 / 8.0)
	} else {
		d.Vk = math.Sqrt(2.0 / 8.0)
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			d.Const_matrix[i][j] = math.Cos((math.Pi*float64((2*j)+1)*float64(u))/16.0) * math.Cos((math.Pi*float64((2*i)+1)*float64(v))/16.0)
		}
	}
	d.Nc = d.Uk * d.Vk

	return d
}

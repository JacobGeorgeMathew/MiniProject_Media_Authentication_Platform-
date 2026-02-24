package watermark

import (
	"math"
)

type Coefficients struct {
	U            int
	V            int
	Const_matrix [32][32]float64
	Uk           float64
	Vk           float64
	Nc           float64
}

func (d *Coefficients) FindVector(block [][]float64) float64 {

	// Calculate DCT coefficient for a 32x32 block
	total := 0.0
	for y := 0; y < 32; y++ { // Fixed: was 8, should be 32 to match block size
		for x := 0; x < 32; x++ { // Fixed: was 8, should be 32 to match block size
			total += block[x][y] * d.Const_matrix[y][x]
		}
	}

	return d.Nc * total
}

func CalculateConstant(h int, w int) []Coefficients { // Fixed: return []Coefficients, not *[]Coefficients
	dct_matrices := make([]Coefficients, 0)
	for v := 0; v < h; v++ {

		for u := 0; u < w; u++ {
			d := new(Coefficients)

			d.U = u
			d.V = v
			// Pre-calculate u and v normalization coefficients
			if u == 0 {
				d.Uk = math.Sqrt(1.0 / 32.0) // Fixed: normalise for N=32
			} else {
				d.Uk = math.Sqrt(2.0 / 32.0) // Fixed: normalise for N=32
			}

			if v == 0 {
				d.Vk = math.Sqrt(1.0 / 32.0) // Fixed: normalise for N=32
			} else {
				d.Vk = math.Sqrt(2.0 / 32.0) // Fixed: normalise for N=32
			}

			for i := 0; i < 32; i++ {
				for j := 0; j < 32; j++ {
					// Fixed: denominator should be 2*N = 64 for a 32-point DCT
					d.Const_matrix[i][j] = math.Cos((math.Pi*float64((2*j)+1)*float64(u))/64.0) *
						math.Cos((math.Pi*float64((2*i)+1)*float64(v))/64.0)
				}
			}
			d.Nc = d.Uk * d.Vk

			dct_matrices = append(dct_matrices, *d)
		}
	}
	return dct_matrices // Fixed: return value directly, not pointer
}
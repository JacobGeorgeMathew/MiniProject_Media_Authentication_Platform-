// package fingerprint

// import (
// 	//"fmt"
// 	"image"

// 	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/engine"
// 	"golang.org/x/image/draw"
// )

// func GetBlock(matrix [][]float64, x, y, B int) [][]float64 {
// 	block := make([][]float64, B)
// 	for i := 0; i < B; i++ {
// 		block[i] = make([]float64, B)
// 		copy(block[i], matrix[y+i][x:x+B])
// 	}
// 	return block
// }

// func PutBlock(matrix [][]float64, block [][]float64, x, y int) {
// 	for i := 0; i < len(block); i++ {
// 		copy(matrix[y+i][x:x+len(block)], block[i])
// 	}
// }

// func ResizeImage(img image.Image, width, height int) *image.RGBA {
// 	dst := image.NewRGBA(image.Rect(0, 0, width, height))
// 	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
// 	return dst
// }

// func Createfingerprint(img image.Image) []float64 { // Fixed: return the vector
// 	const_matrices := CalculateConstant(4, 4) // Fixed: no longer a pointer, used directly

// 	resized_img := ResizeImage(img, 256, 256)

// 	_, Ymatrix := engine.ConvertToYC(resized_img)

// 	vector1024d := []float64{} // Fixed: missing {} for slice literal

// 	for i := 0; i < 8; i++ {
// 		for j := 0; j < 8; j++ {
// 			block := GetBlock(Ymatrix, i*32, j*32, 32)

// 			vector16d := []float64{} // Fixed: missing {} for slice literal

// 			for c := 0; c < 16; c++ {
// 				// Fixed: const_matrices is no longer a pointer, no need to dereference
// 				vector16d = append(vector16d, const_matrices[c].FindVector(block))
// 			}
// 			vector1024d = append(vector1024d, vector16d...)
// 		}
// 	}

// 	//fmt.Println(vector1024d) // Fixed: use fmt.Println instead of println for slices
// 	return vector1024d // Fixed: return the fingerprint vector
// }


package fingerprint

import (
	"fmt"
	"math"
	"image"
)

const (
	SIZE       = 256
)
// ─────────────────────────────────────────────
//  DWT FILTERS  (Daubechies db4)
// ─────────────────────────────────────────────

var hFilt = []float64{0.4829629131, 0.8365163037, 0.2241438680, -0.1294095226}
var gFilt = []float64{-0.1294095226, -0.2241438680, 0.8365163037, -0.4829629131}

func convertToY(img image.Image) [][]float64 {
	b := img.Bounds()
	m := make([][]float64, b.Dy())
	for i := range m {
		m[i] = make([]float64, b.Dx())
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bv, _ := img.At(x, y).RGBA()
			Y := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bv>>8)
			m[y-b.Min.Y][x-b.Min.X] = Y - 128.0
		}
	}
	return m
}

func resizeNearest(src [][]float64, newH, newW int) [][]float64 {
	h, w := len(src), len(src[0])
	dst := make([][]float64, newH)
	for i := range dst {
		dst[i] = make([]float64, newW)
		for j := 0; j < newW; j++ {
			dst[i][j] = src[i*h/newH][j*w/newW]
		}
	}
	return dst
}

func mkMat(r, c int) [][]float64 {
	m := make([][]float64, r)
	for i := range m {
		m[i] = make([]float64, c)
	}
	return m
}

func blockMaxVector(mat [][]float64) []float64 {
	size, block := len(mat), 8
	vec := make([]float64, 0, (size/block)*(size/block))
	for i := 0; i < size; i += block {
		for j := 0; j < size; j += block {
			maxV := 0.0
			for y := i; y < i+block; y++ {
				for x := j; x < j+block; x++ {
					if v := math.Abs(mat[y][x]); v > maxV {
						maxV = v
					}
				}
			}
			vec = append(vec, maxV)
		}
	}
	return vec
}

func normalize(v []float64) []float64 {
	var norm float64
	for _, x := range v {
		norm += x * x
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return v
	}
	for i := range v {
		v[i] /= norm
	}
	return v
}

func dwt1D(sig []float64) ([]float64, []float64) {
	n := len(sig)
	half := n / 2
	low, high := make([]float64, half), make([]float64, half)
	for i := 0; i < half; i++ {
		for k := 0; k < len(hFilt); k++ {
			idx := (2*i + k) % n
			low[i] += sig[idx] * hFilt[k]
			high[i] += sig[idx] * gFilt[k]
		}
	}
	return low, high
}

func dwt2D(mat [][]float64) ([][]float64, [][]float64, [][]float64, [][]float64) {
	rows, cols := len(mat), len(mat[0])
	lowR := make([][]float64, rows)
	highR := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		lowR[i], highR[i] = dwt1D(mat[i])
	}
	hR, hC := rows/2, cols/2
	ll, lh, hl, hh := mkMat(hR, hC), mkMat(hR, hC), mkMat(hR, hC), mkMat(hR, hC)
	for j := 0; j < hC; j++ {
		cL, cH := make([]float64, rows), make([]float64, rows)
		for i := 0; i < rows; i++ {
			cL[i], cH[i] = lowR[i][j], highR[i][j]
		}
		ll_, lh_ := dwt1D(cL)
		hl_, hh_ := dwt1D(cH)
		for i := 0; i < hR; i++ {
			ll[i][j], lh[i][j], hl[i][j], hh[i][j] = ll_[i], lh_[i], hl_[i], hh_[i]
		}
	}
	return ll, lh, hl, hh
}

func imageToVector(img image.Image) ([]float64, error) {
	

	yMat := convertToY(img)
	yMat = resizeNearest(yMat, SIZE, SIZE)

	ll0, _, _, _ := dwt2D(yMat)
	ll1, lh1, hl1, hh1 := dwt2D(ll0)

	vec := make([]float64, 0, 256)
	vec = append(vec, blockMaxVector(ll1)...)
	vec = append(vec, blockMaxVector(lh1)...)
	vec = append(vec, blockMaxVector(hl1)...)
	vec = append(vec, blockMaxVector(hh1)...)

	vec = normalize(vec)
	return vec, nil
}

func Createfingerprint(img image.Image) []float64 {

	vec , err := imageToVector(img)

	if err != nil {
			fmt.Printf("  [FAIL] Vector error : %v\n", err)
			
		}

	return  vec
}
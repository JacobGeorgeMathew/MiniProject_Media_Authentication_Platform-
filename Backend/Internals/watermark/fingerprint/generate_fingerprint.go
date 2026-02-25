package fingerprint

import (
	//"fmt"
	"image"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/engine"
	"golang.org/x/image/draw"
)

func GetBlock(matrix [][]float64, x, y, B int) [][]float64 {
	block := make([][]float64, B)
	for i := 0; i < B; i++ {
		block[i] = make([]float64, B)
		copy(block[i], matrix[y+i][x:x+B])
	}
	return block
}

func PutBlock(matrix [][]float64, block [][]float64, x, y int) {
	for i := 0; i < len(block); i++ {
		copy(matrix[y+i][x:x+len(block)], block[i])
	}
}

func ResizeImage(img image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

func Createfingerprint(img image.Image) []float64 { // Fixed: return the vector
	const_matrices := CalculateConstant(4, 4) // Fixed: no longer a pointer, used directly

	resized_img := ResizeImage(img, 256, 256)

	_, Ymatrix := engine.ConvertToYC(resized_img)

	vector1024d := []float64{} // Fixed: missing {} for slice literal

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			block := GetBlock(Ymatrix, i*32, j*32, 32)

			vector16d := []float64{} // Fixed: missing {} for slice literal

			for c := 0; c < 16; c++ {
				// Fixed: const_matrices is no longer a pointer, no need to dereference
				vector16d = append(vector16d, const_matrices[c].FindVector(block))
			}
			vector1024d = append(vector1024d, vector16d...)
		}
	}

	//fmt.Println(vector1024d) // Fixed: use fmt.Println instead of println for slices
	return vector1024d // Fixed: return the fingerprint vector
}

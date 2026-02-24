package watermarkengine


import (
	"fmt"
	"image"
	"math"
)

func ConvertToYC(img image.Image) (*image.YCbCr, [][]float64) {
	bounds := img.Bounds()

	// Create a new YCbCr image with the same size
	ycb := image.NewYCbCr(bounds, image.YCbCrSubsampleRatio444)

	Ymatrix := make([][]float64, bounds.Dy())
	for i := range Ymatrix {
		Ymatrix[i] = make([]float64, bounds.Dx())
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert using the standard Go formula
			rr := float64(r >> 8)
			gg := float64(g >> 8)
			bb := float64(b >> 8)

			Y := 0.299*rr + 0.587*gg + 0.114*bb
			Cb := -0.1687*rr - 0.3313*gg + 0.5*bb + 128
			Cr := 0.5*rr - 0.4187*gg - 0.0813*bb + 128

			yi := y - bounds.Min.Y
			xi := x - bounds.Min.X

			// Store Y component normalized by subtracting 128 (centered at 0)
			Ymatrix[yi][xi] = Y - 128.0
			ycb.Y[ycb.YOffset(x, y)] = uint8(Y)
			ycb.Cb[ycb.COffset(x, y)] = uint8(Cb)
			ycb.Cr[ycb.COffset(x, y)] = uint8(Cr)
		}
	}
	return ycb, Ymatrix
}

func Modify_YComponent(ycb *image.YCbCr, Ymatrix [][]float64) {
	bounds := ycb.Bounds()

	// First pass: find the actual range of values
	minValue := Ymatrix[0][0]
	maxValue := Ymatrix[0][0]

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			val := Ymatrix[y][x]
			if val < minValue {
				minValue = val
			}
			if val > maxValue {
				maxValue = val
			}
		}
	}

	fmt.Printf("Before normalization - Min: %.2f, Max: %.2f\n", minValue, maxValue)

	// Check if values are within expected range [-128, 127]
	needsNormalization := minValue < -128 || maxValue > 127

	if needsNormalization {
		// Adaptive normalization: scale to fit within [-128, 127] range
		valueRange := maxValue - minValue
		targetRange := 255.0 // Range from -128 to 127

		fmt.Printf("Applying adaptive normalization (range: %.2f -> %.2f)\n", valueRange, targetRange)
		h := int (math.Floor(float64(bounds.Dy() / 256)) * 256 )
		w := int (math.Floor(float64(bounds.Dx() / 256)) * 256 )
		for y := bounds.Min.Y; y < h; y++ {
			for x := bounds.Min.X; x < w; x++ {
				yi := y - bounds.Min.Y
				xi := x - bounds.Min.X

				// Normalize to [-128, 127] range, then shift to [0, 255]
				normalized := ((Ymatrix[yi][xi]-minValue)/valueRange)*targetRange - 128.0
				value := normalized + 128.0

				ycb.Y[ycb.YOffset(x, y)] = uint8(math.Min(math.Max(value, 0), 255))
			}
		}
	} else {
		// Values are within range, just add 128
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				yi := y - bounds.Min.Y
				xi := x - bounds.Min.X

				value := Ymatrix[yi][xi] + 128.0
				ycb.Y[ycb.YOffset(x, y)] = uint8(math.Min(math.Max(value, 0), 255))
			}
		}
	}

	fmt.Println("Y component modification complete")
}

package engine

import (
	"fmt"
	"image"
)

func ExtractWatermark(img image.Image, c []Constants) ([][]int, bool) {
	// Convert image to YCbCr and get Y matrix
	_, Ymatrix := ConvertToYC(img)

	fmt.Println("Extraction started")
	fmt.Printf("Image Y matrix dimensions: %dx%d\n", len(Ymatrix[0]), len(Ymatrix))

	x_index, y_index, _ := Identify(img, c)

	fmt.Println("X_index : ", x_index, "Y_index : ", y_index)

	h := len(Ymatrix)
	w := len(Ymatrix[0])

	//var messages []string
	tileCount := 0
	// validTileCount := 0

	// if !flag {
	// 	return messages, false
	// }

	// Process each 256x256 tile
	numTilesY := h / 256
	numTilesX := w / 256

	fmt.Printf("Processing %d x %d = %d tiles\n", numTilesY, numTilesX, numTilesY*numTilesX)
	fmt.Println("----------------------------------------")
	extractedBits := make([][]int, numTilesX*numTilesY)
	for i := x_index; i < numTilesY; i++ {
		for j := y_index; j < numTilesX; j++ {

			// Get the tile from the Y matrix (not DWT transformed)
			tile := GetBlock(Ymatrix, j*256, i*256, 256)

			// Extract bits from this tile (DWT happens inside ExtractfromaTile)
			extractedBits[tileCount] = ExtractfromaTile(tile, c)

			fmt.Println("Length of one payload : ", len(extractedBits[tileCount]))
			
			tileCount++
		}
	}
	return extractedBits, true
}


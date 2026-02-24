package watermarkengine

import (
	"fmt"
	"image"
	"math"
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

func EmbedWatermark(img image.Image, payload []int, c []Constants) (*image.YCbCr , bool) {
	stream := payload

	//fmt.Printf("payload: \"%s\" -> %d bits (including flags)\n", payload, len(stream))

	ycb, Ymatrix := ConvertToYC(img)

	fmt.Printf("Image converted to YCbCr, Y matrix size: %dx%d\n", len(Ymatrix[0]), len(Ymatrix))
	_ , _ , flag := Identify(Ymatrix,c)
	if flag {
		return ycb , true
	}
	h := len(Ymatrix)
	w := len(Ymatrix[0])

	// Calculate number of tiles
	numTilesY := int(math.Floor(float64(h) / 256))
	numTilesX := int(math.Floor(float64(w) / 256))

	fmt.Printf("Processing %d x %d = %d tiles\n", numTilesY, numTilesX, numTilesY*numTilesX)

	// Calculate capacity per tile
	// Each tile is 256x256, divided into 16x16 blocks = 8x8 = 64 blocks
	// First row (8 blocks) + first column (7 blocks, excluding corner) = 15 blocks for verification
	// Remaining: 64 - 15 = 49 blocks for data
	blocksPerTile := 49
	bitsPerTile := blocksPerTile * 2

	fmt.Printf("Capacity per tile: %d bits\n", bitsPerTile)
	fmt.Printf("Total capacity: %d bits\n", bitsPerTile*numTilesY*numTilesX)

	// if len(stream) > bitsPerTile {
	// 	fmt.Printf("⚠ Warning: payload (%d bits) exceeds single tile capacity (%d bits)\n",
	// 		len(stream), bitsPerTile)
	// 	fmt.Printf("   payload will be replicated across multiple tiles\n")
	// }

	// Process tiles
	tileCount := 0
	for i := 0; i < numTilesY; i++ {
		for j := 0; j < numTilesX; j++ {
			tileCount++
			
				// Get the tile from the Y matrix (spatial domain, not DWT yet)
				tile := GetBlock(Ymatrix, j*256, i*256, 256)

				// Embed watermark in this tile
				// DWT will be performed inside EmbedinaTile on 16x16 blocks
				modifiedTile := EmbedinaTile(tile, stream, c)

				// Put the modified tile back
				PutBlock(Ymatrix, modifiedTile, j*256, i*256)
			
			fmt.Printf("✓ Tile [%d,%d] (tile #%d): Watermark embedded\n", i, j, tileCount)
		}
	}

	// Update the Y component with the modified matrix
	Modify_YComponent(ycb, Ymatrix)

	fmt.Println("Watermark embedding completed successfully")

	return ycb , false
}

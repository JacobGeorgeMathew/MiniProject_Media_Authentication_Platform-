package watermarkengine

import (
	"math"
)

func qimembed(c float64, bit int, delta float64) float64 {
	base := math.Floor(c/delta) * delta

	if bit == 0 {
		// Quantize to 0*delta + delta/4
		return base + delta/4
	}
	// Quantize to 0*delta + 3*delta/4
	return base + 3*delta/4
}

func qimExtract(c float64, delta float64) int {
	base := math.Floor(c/delta) * delta
	remainder := c - base

	if remainder < delta/2 {
		return 0
	}
	return 1
}

func PerformEmbed(block [][]float64, bits []int, c []Constants) {
	alpha := 50.0

	// Calculate current DCT coefficient
	coff1 := c[0].FindValueOptimized(block)
	coff2 := c[1].FindValueOptimized(block)

	// Calculate target quantized coefficient value
	quantizedCoff1 := qimembed(coff1, bits[0], alpha)
	quantizedCoff2 := qimembed(coff2, bits[1], alpha)
	// Calculate the change needed
	delta_coff1 := quantizedCoff1 - coff1
	delta_coff2 := quantizedCoff2 - coff2
	// Distribute the change back to spatial domain
	// The coefficient change delta_coff needs to be distributed
	// using the basis function, weighted by Nc
	multiplier1 := delta_coff1 * c[0].Nc
	multiplier2 := delta_coff2 * c[1].Nc

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			block[x][y] += (multiplier1*c[0].Const_matrix[y][x] + multiplier2*c[1].Const_matrix[y][x])
		}
	}
}

func PerformExtract(block [][]float64, c *Constants) int {
	alpha := 50.0

	// Calculate DCT coefficient
	coff := c.FindValueOptimized(block)

	// Extract bit using QIM
	bit := qimExtract(coff, alpha)

	return bit
}

// markatile marks the first row and first column of a tile with bit "1"
// This creates a verification pattern: all blocks at (0, bx) and (by, 0) contain bit 1
func markatile(tile [][]float64, c []Constants) [][]float64 {
	bits := make([]int, 2)
	bits[0] = 1
	bits[1] = 1
	// Mark the first row (by = 0, bx = 0, 16, 32, ..., 112)
	by := 0
	for bx := 0; bx < 256; bx += 16 {
		block := GetBlock(tile, bx, by, 16)
		block_DWT := PerformCompleteDWT(block)

		// Embed bit 1 in the HL component
		PerformEmbed(block_DWT.HL, bits, c)

		modified_block := PerformCompleteIDWT(block_DWT.LL, block_DWT.LH, block_DWT.HL, block_DWT.HH)
		// Put the modified block back
		PutBlock(tile, modified_block, bx, by)
	}

	// Mark the first column (bx = 0, by = 16, 32, ..., 112)
	// Note: (0,0) is already marked above, so start from by=16
	bx := 0
	for by = 16; by < 256; by += 16 {
		block := GetBlock(tile, bx, by, 16)
		block_DWT := PerformCompleteDWT(block)

		// Embed bit 1 in the HL component
		PerformEmbed(block_DWT.HL, bits, c)

		modified_block := PerformCompleteIDWT(block_DWT.LL, block_DWT.LH, block_DWT.HL, block_DWT.HH)
		// Put the modified block back
		PutBlock(tile, modified_block, bx, by)
	}

	return tile
}

func EmbedinaTile(tile [][]float64, stream []int, c []Constants) [][]float64 {
	bitIndex := 0
	bits := make([]int, 2)
	// First, mark the tile with verification pattern

	tile = markatile(tile, c)

	// Then embed the actual watermark data in the remaining blocks
	// Skip first row and first column (they contain the verification pattern)
	for by := 16; by < 256; by += 16 {
		for bx := 16; bx < 256; bx += 16 {

			block := GetBlock(tile, bx, by, 16)

			block_DWT := PerformCompleteDWT(block)
			if bitIndex < len(stream) {
				bits[0] = stream[bitIndex]
				bits[1] = stream[bitIndex+1]
				// Embed the bit in the HL component
				PerformEmbed(block_DWT.HL, bits, c)
				bitIndex += 2
			}
			modified_block := PerformCompleteIDWT(block_DWT.LL, block_DWT.LH, block_DWT.HL, block_DWT.HH)
			// Put the modified block back
			PutBlock(tile, modified_block, bx, by)

		}
	}

	return tile
}

// verifytile verifies that a tile contains the expected verification pattern
// Returns true if the first row and first column all contain bit "1"
func Verifytile(tile [][]float64, c []Constants, flag bool) bool {
	correctBits := 0
	totalBits := 0
	if flag {
		// Check the first row (by = 0, bx = 0, 16, 32, ..., 240)
		by := 0
		for bx := 0; bx < 256; bx += 16 { // was 512, should be 256
			block := GetBlock(tile, bx, by, 16)
			block_DWT := PerformCompleteDWT(block)

			for d := range c {
				bit := PerformExtract(block_DWT.HL, &c[d])
				totalBits++
				if bit == 1 {
					correctBits++
				}
			}
		}
	} else {
		// Check the first column (bx = 0, by = 16, 32, ..., 240)
		bx := 0
		for by := 16; by < 256; by += 16 {
			block := GetBlock(tile, bx, by, 16)
			block_DWT := PerformCompleteDWT(block)

			for d := range c {
				bit := PerformExtract(block_DWT.HL, &c[d])
				totalBits++
				if bit == 1 {
					correctBits++
				}
			}
		}
	}

	successRate := float64(correctBits) / float64(totalBits)
	threshold := 0.7

	return successRate >= threshold
}

func ExtractfromaTile(tile [][]float64, c []Constants) []int {
	var extractedBits []int

	// Verify the tile first
	//tile_status := verifytile(tile, c)

	// Extract data from the non-verification blocks
	// Skip first row and first column (they contain the verification pattern)
	for by := 16; by < 256; by += 16 {
		for bx := 16; bx < 256; bx += 16 {
			block := GetBlock(tile, bx, by, 16)
			block_DWT := PerformCompleteDWT(block)

			// Extract bit from HL component
			for d := range c {
				bit := PerformExtract(block_DWT.HL, &c[d])

				extractedBits = append(extractedBits, bit)
			}
		}
	}

	return extractedBits
}

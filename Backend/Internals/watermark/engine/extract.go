package engine

import (
	"fmt"
	"image"
)

// findMessage locates the message between start and end flags
// func findMessage(bits []int) (string, bool) {
// 	startFlag := []int{
// 		1, 1, 1, 1, 0, 0, 0, 0,
// 		1, 1, 1, 1, 0, 0, 0, 0,
// 	}
// 	endFlag := []int{
// 		0, 0, 0, 0, 1, 1, 1, 1,
// 		0, 0, 0, 0, 1, 1, 1, 1,
// 	}

// 	// Find start flag
// 	startIndex := -1
// 	for i := 0; i <= len(bits)-len(startFlag); i++ {
// 		match := true
// 		for j := 0; j < len(startFlag); j++ {
// 			if bits[i+j] != startFlag[j] {
// 				match = false
// 				break
// 			}
// 		}
// 		if match {
// 			startIndex = i + len(startFlag)
// 			break
// 		}
// 	}

// 	if startIndex == -1 {
// 		return "", false
// 	}

// 	// Find end flag after start
// 	endIndex := -1
// 	for i := startIndex; i <= len(bits)-len(endFlag); i++ {
// 		match := true
// 		for j := 0; j < len(endFlag); j++ {
// 			if bits[i+j] != endFlag[j] {
// 				match = false
// 				break
// 			}
// 		}
// 		if match {
// 			endIndex = i
// 			break
// 		}
// 	}

// 	if endIndex == -1 {
// 		return "", false
// 	}

// 	// Extract message bits between flags
// 	messageBits := bits[startIndex:endIndex]

// 	// Convert bits to bytes
// 	if len(messageBits)%8 != 0 {
// 		// Pad with zeros if needed
// 		padding := 8 - (len(messageBits) % 8)
// 		messageBits = append(messageBits, make([]int, padding)...)
// 	}

// 	messageBytes := bitsToBytes(messageBits)
// 	return string(messageBytes), true
// }

func ExtractWatermark(img image.Image, c []Constants) ([][]int, bool) {
	// Convert image to YCbCr and get Y matrix
	_, Ymatrix := ConvertToYC(img)

	fmt.Println("Extraction started")
	fmt.Printf("Image Y matrix dimensions: %dx%d\n", len(Ymatrix[0]), len(Ymatrix))

	x_index, y_index, _ := Identify(Ymatrix, c)

	fmt.Println("X_index : ",x_index, "Y_index : ", y_index)

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
	extractedBits := make([][]int,numTilesX * numTilesY)
	for i := x_index; i < numTilesY; i++ {
		for j := y_index; j < numTilesX; j++ {
			

			// Get the tile from the Y matrix (not DWT transformed)
			tile := GetBlock(Ymatrix, j*256, i*256, 256)

			// Extract bits from this tile (DWT happens inside ExtractfromaTile)
			copy(extractedBits[tileCount],ExtractfromaTile(tile, c))
			// tile_status := true
			// // if tile_status {
			// // 	validTileCount++
			// // }

			// // Try to find the message
			// message, found := findMessage(extractedBits)

			// if found && tile_status {
			// 	fmt.Printf("✓ Tile [%d,%d] (tile #%d): VERIFIED & Message found: \"%s\"\n",
			// 		i, j, tileCount, message)
			// 	messages = append(messages, message)
			// } else if found && !tile_status {
			// 	fmt.Printf("⚠ Tile [%d,%d] (tile #%d): NOT VERIFIED but message found: \"%s\"\n",
			// 		i, j, tileCount, message)
			// 	// Optionally still add the message with a warning
			// 	messages = append(messages, message)
			// } else if !found && tile_status {
			// 	fmt.Printf("⚠ Tile [%d,%d] (tile #%d): VERIFIED but no valid message\n",
			// 		i, j, tileCount)
			// } else {
			// 	fmt.Printf("✗ Tile [%d,%d] (tile #%d): NOT VERIFIED & no message\n",
			// 		i, j, tileCount)
			// }
			tileCount++
		}
	}

	// fmt.Println("----------------------------------------")
	// fmt.Printf("Summary: %d/%d tiles verified successfully (%.1f%%)\n",
	// 	validTileCount, tileCount, float64(validTileCount)*100.0/float64(tileCount))
	// fmt.Printf("Total messages found: %d\n", len(messages))

	return extractedBits, true
}

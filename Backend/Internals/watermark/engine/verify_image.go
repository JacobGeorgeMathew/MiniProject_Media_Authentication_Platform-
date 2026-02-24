package watermark

func Identify(Ymatrix [][]float64, c []Constants) (x int, y int, flag bool) {
	h := len(Ymatrix)
	w := len(Ymatrix[0])

	// Find x_index: scan candidate tile row-starts (0, 256, 512, ...)
	// For each candidate, take the first 16 rows of that tile and check
	// that the first column (bx=0) verification pattern holds (flag=false).
	x_index := -1
	for i := 0; i+256 <= h; i += 256 {
		// Build a 256x256 block starting at row i, col 0
		block := make([][]float64, 256)
		for row := 0; row < 256; row++ {
			block[row] = make([]float64, 256)
			copy(block[row], Ymatrix[i+row][:256])
		}
		if Verifytile(block, c, false) {
			x_index = i / 256
			break
		}
	}
	if x_index == -1 {
		return -1, -1, false
	}

	// Find y_index: scan candidate tile column-starts (0, 256, 512, ...)
	// For each candidate, take the 256x256 block at (x_index*256, j) and
	// check that the first row verification pattern holds (flag=true).
	y_index := -1
	for j := 0; j+256 <= w; j += 256 {
		block := make([][]float64, 256)
		for row := 0; row < 256; row++ {
			block[row] = make([]float64, 256)
			copy(block[row], Ymatrix[x_index*256+row][j:j+256])
		}
		if Verifytile(block, c, true) {
			y_index = j / 256
			break
		}
	}
	if y_index == -1 {
		return -1, -1, false
	}

	return x_index, y_index, true
}

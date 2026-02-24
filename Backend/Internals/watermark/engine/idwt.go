package watermark

import (
	"fmt"
	"math"
	"sync"
	//"time"
)

// invHaarCombine combines low-pass and high-pass components
// Inverse Haar transform: reconstruct from L and H coefficients
func invHaarCombine(low []float64, high []float64) []float64 {
	n := len(low)
	output := make([]float64, n*2)

	for i := 0; i < n; i++ {
		// Haar forward: L = (x + y)/√2, H = (x - y)/√2
		// Inverse: x = (L + H), y = (L - H)
		// Since L and H are already divided by √2, we need:
		// x = L + H, y = L - H
		output[2*i] = low[i] + high[i]
		output[2*i+1] = low[i] - high[i]
	}
	return output
}

// PerformCompleteIDWT performs inverse 2D DWT using all four components
// This provides perfect reconstruction of the original matrix
func PerformCompleteIDWT(LL, LH, HL, HH [][]float64) [][]float64 {
	//t1 := time.Now()

	h := len(LL)
	w := len(LL[0])

	// Validate all components have same dimensions
	if len(LH) != h || len(HL) != h || len(HH) != h ||
		len(LH[0]) != w || len(HL[0]) != w || len(HH[0]) != w {
		panic("All DWT components must have the same dimensions")
	}

	var wg sync.WaitGroup

	// Step 1: Inverse column-wise transform
	// Combine LL with LH to get tempL (low-pass rows)
	tempL := make([][]float64, h*2)
	for i := range tempL {
		tempL[i] = make([]float64, w)
	}

	// Combine HL with HH to get tempH (high-pass rows)
	tempH := make([][]float64, h*2)
	for i := range tempH {
		tempH[i] = make([]float64, w)
	}

	// Process each column in parallel
	for col := 0; col < w; col++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()

			// Extract columns from LL and LH
			lowColumn := make([]float64, h)
			highColumn := make([]float64, h)
			for row := 0; row < h; row++ {
				lowColumn[row] = LL[row][c]
				highColumn[row] = LH[row][c]
			}

			// Inverse transform to get tempL column
			reconstructedL := invHaarCombine(lowColumn, highColumn)
			for row := 0; row < h*2; row++ {
				tempL[row][c] = reconstructedL[row]
			}

			// Extract columns from HL and HH
			for row := 0; row < h; row++ {
				lowColumn[row] = HL[row][c]
				highColumn[row] = HH[row][c]
			}

			// Inverse transform to get tempH column
			reconstructedH := invHaarCombine(lowColumn, highColumn)
			for row := 0; row < h*2; row++ {
				tempH[row][c] = reconstructedH[row]
			}
		}(col)
	}
	wg.Wait()

	// Step 2: Inverse row-wise transform
	// Combine tempL and tempH to get final result
	result := make([][]float64, h*2)
	for i := range result {
		result[i] = make([]float64, w*2)
	}

	for row := 0; row < h*2; row++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()

			// Combine tempL and tempH rows
			reconstructed := invHaarCombine(tempL[r], tempH[r])

			for col := 0; col < w*2; col++ {
				result[r][col] = reconstructed[col]
			}
		}(row)
	}
	wg.Wait()

	//t2 := time.Now()
	//fmt.Printf("Inverse DWT completed in %v\n", t2.Sub(t1))

	return result
}

// PerformCompleteIDWTFromResult performs inverse DWT from DWTResult struct
func PerformCompleteIDWTFromResult(dwtResult *DWTResult) [][]float64 {
	return PerformCompleteIDWT(dwtResult.LL, dwtResult.LH, dwtResult.HL, dwtResult.HH)
}

// CalculateReconstructionError computes error metrics between original and reconstructed
func CalculateReconstructionError(original, reconstructed [][]float64) {
	h := len(original)
	w := len(original[0])

	if len(reconstructed) != h || len(reconstructed[0]) != w {
		fmt.Printf("ERROR: Dimension mismatch - Original: %dx%d, Reconstructed: %dx%d\n",
			h, w, len(reconstructed), len(reconstructed[0]))
		return
	}

	var sumSquaredError float64
	var sumAbsError float64
	var maxError float64
	count := 0

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			error := math.Abs(original[i][j] - reconstructed[i][j])
			sumAbsError += error
			sumSquaredError += error * error
			if error > maxError {
				maxError = error
			}
			count++
		}
	}

	mae := sumAbsError / float64(count)
	mse := sumSquaredError / float64(count)
	rmse := math.Sqrt(mse)

	fmt.Println("\n=== Reconstruction Error Metrics ===")
	fmt.Printf("Mean Absolute Error (MAE):     %.10f\n", mae)
	fmt.Printf("Mean Squared Error (MSE):      %.10f\n", mse)
	fmt.Printf("Root Mean Squared Error (RMSE): %.10f\n", rmse)
	fmt.Printf("Max Absolute Error:            %.10f\n", maxError)

	if maxError < 1e-10 {
		fmt.Println("✓ Perfect reconstruction achieved!")
	} else if maxError < 1e-6 {
		fmt.Println("✓ Excellent reconstruction (numerical precision)")
	} else {
		fmt.Println("⚠ Reconstruction has noticeable errors")
	}
}

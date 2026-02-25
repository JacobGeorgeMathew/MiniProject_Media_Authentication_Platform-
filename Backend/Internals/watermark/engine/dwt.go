package engine

import (
	"fmt"
	"math"
	"sync"
	//"time"
)

// haarL performs Haar low-pass filter (averaging)
func haarL(input []float64) []float64 {
	n := len(input)
	output := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		output[i] = (input[2*i] + input[2*i+1]) / math.Sqrt2
	}
	return output
}

// haarH performs Haar high-pass filter (differencing)
func haarH(input []float64) []float64 {
	n := len(input)
	output := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		output[i] = (input[2*i] - input[2*i+1]) / math.Sqrt2
	}
	return output
}

// DWTResult holds all four components of 2D DWT
type DWTResult struct {
	LL [][]float64 // Low-Low (Approximation)
	LH [][]float64 // Low-High (Horizontal details)
	HL [][]float64 // High-Low (Vertical details)
	HH [][]float64 // High-High (Diagonal details)
}

// PerformCompleteDWT performs 2D DWT and returns all four sub-bands
func PerformCompleteDWT(Ymatrix [][]float64) *DWTResult {
	//t1 := time.Now()

	h := len(Ymatrix)
	w := len(Ymatrix[0])

	// Step 1: Row-wise transform (both L and H)
	tempL := make([][]float64, h) // Low-pass on rows
	tempH := make([][]float64, h) // High-pass on rows

	var wg sync.WaitGroup

	// Process each row in parallel
	for i := 0; i < h; i++ {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			tempL[row] = haarL(Ymatrix[row])
			tempH[row] = haarH(Ymatrix[row])
		}(i)
	}
	wg.Wait()

	// Step 2: Column-wise transform on tempL to get LL and LH
	LL := make([][]float64, h/2)
	LH := make([][]float64, h/2)
	for i := range LL {
		LL[i] = make([]float64, w/2)
		LH[i] = make([]float64, w/2)
	}

	for col := 0; col < w/2; col++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()

			// Extract column from tempL
			column := make([]float64, h)
			for row := 0; row < h; row++ {
				column[row] = tempL[row][c]
			}

			// Apply L and H transforms
			columnL := haarL(column)
			columnH := haarH(column)

			// Store results
			for row := 0; row < h/2; row++ {
				LL[row][c] = columnL[row]
				LH[row][c] = columnH[row]
			}
		}(col)
	}
	wg.Wait()

	// Step 3: Column-wise transform on tempH to get HL and HH
	HL := make([][]float64, h/2)
	HH := make([][]float64, h/2)
	for i := range HL {
		HL[i] = make([]float64, w/2)
		HH[i] = make([]float64, w/2)
	}

	for col := 0; col < w/2; col++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()

			// Extract column from tempH
			column := make([]float64, h)
			for row := 0; row < h; row++ {
				column[row] = tempH[row][c]
			}

			// Apply L and H transforms
			columnL := haarL(column)
			columnH := haarH(column)

			// Store results
			for row := 0; row < h/2; row++ {
				HL[row][c] = columnL[row]
				HH[row][c] = columnH[row]
			}
		}(col)
	}
	wg.Wait()

	//t2 := time.Now()
	//fmt.Printf("DWT completed in %v\n", t2.Sub(t1))

	return &DWTResult{
		LL: LL,
		LH: LH,
		HL: HL,
		HH: HH,
	}
}

// GetStatistics computes min, max, and range for a 2D matrix
func GetStatistics(matrix [][]float64, name string) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return
	}

	min := matrix[0][0]
	max := matrix[0][0]

	for _, row := range matrix {
		for _, val := range row {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	fmt.Printf("%s - Min: %.4f, Max: %.4f, Range: %.4f\n", name, min, max, max-min)
}

// PrintDWTStatistics prints statistics for all DWT components
func PrintDWTStatistics(result *DWTResult) {
	fmt.Println("\n=== DWT Component Statistics ===")
	GetStatistics(result.LL, "LL (Approximation)")
	GetStatistics(result.LH, "LH (Horizontal)")
	GetStatistics(result.HL, "HL (Vertical)")
	GetStatistics(result.HH, "HH (Diagonal)")
}

/*
 Copyright (C) 2022 Douglas Wayne Potter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// naive matrix math code. It partially exploits sparsity, but lacks
// other obvious optimizations matrix blocking, SIMD, optimized
// memory access patterns, etc. However if matrices are banded or mostly
// banded, this code should decently performant.

package math

import (
	"errors"
	"fmt"
	"math"
)

// sen is a special value indicating the end of the row/column.
const sen = math.MaxUint32 // sentential value (= 4294967295)

// cRSMatrix is a Compressed Row Storage Matrix for a binary matrix
// The values are the column indices of 1s, except the sentential value
// which indicate the end of the row.
// For example:
// 0 0 0 1 1 =>  [3, 4, sen, sen, 1, 2, sen]
// 0 0 0 0 0
// 0 1 1 0 0
type cRSMatrix []uint32

// cCSMatrix is a Compressed Column Storage Matrix for a binary matrix
// The values are the row indices of 1s, except the sentential value
// which indicate the end of the column.
// For example:
// 0 0 0  => [3, 4, sen, sen, 1, 2, sen]
// 0 0 1
// 0 0 1
// 1 0 0
// 1 0 0
// (This is transpose of the example cRSMatrix above.)
type cCSMatrix []uint32

func hasIndices(x []uint32) bool {
	foundIndex := false
	for i := 0; i < len(x); i++ {
		if x[i] != sen {
			foundIndex = true
			break
		}
	}
	return foundIndex
}

func checkValidMatrixIndices(x []uint32) error {
	if !hasIndices(x) {
		return errors.New("must contain indices")
	}

	if x[len(x)-1] != sen {
		return fmt.Errorf("must end in the sentential value (%d)", sen)
	}

	return nil
}

func makeCompressedRowMatrix(x []uint32) (cRSMatrix, error) {
	if err := checkValidMatrixIndices(x); err != nil {
		return []uint32{}, err
	}
	return cRSMatrix(x), nil

}

func makeCompressedColumnMatrix(x []uint32) (cCSMatrix, error) {
	if err := checkValidMatrixIndices(x); err != nil {
		return []uint32{}, err
	}

	return cCSMatrix(x), nil
}

func transpose(x []uint32) ([]uint32, error) {
	// The notation (comments and variable names) are as if we are going
	// from row to column representation but the mechanics of the code work
	// to go from colum to row representation as well.

	nElementsPerColumn := make([]uint32, 0)
	nnz := 0
	for i := 0; i < len(x); i++ {
		if x[i] == sen {
			continue
		}
		colIndex := x[i]

		nnz++
		nCols := uint32(len(nElementsPerColumn))
		if colIndex >= nCols {
			nMoreColumns := colIndex - nCols + 1
			for j := uint32(0); j < nMoreColumns; j++ {
				nElementsPerColumn = append(nElementsPerColumn, 0)
			}
		}
		nElementsPerColumn[colIndex]++
	}

	nCols := len(nElementsPerColumn)

	// t = transposed
	t := make([]uint32, nnz+nCols)

	currentColumnPos := make([]uint32, len(nElementsPerColumn))
	for i := 1; i < len(currentColumnPos); i++ {
		// + 1 for for a space for the sentinel
		currentColumnPos[i] = currentColumnPos[i-1] + nElementsPerColumn[i-1] + 1
	}

	// place new column sentinels
	for j := 0; j < len(currentColumnPos); j++ {
		t[int(currentColumnPos[j])+int(nElementsPerColumn[j])] = sen
	}

	// place row positions
	var row uint32
	for i := 0; i < len(x); i++ {
		if x[i] == sen {
			row++
			continue
		}
		if t[currentColumnPos[x[i]]] == sen {
			return []uint32{}, fmt.Errorf("transpose: overwrote sentintel value indicating indexing error")
		}
		t[currentColumnPos[x[i]]] = row
		currentColumnPos[x[i]]++
	}

	if err := checkValidMatrixIndices(t); err != nil {
		return []uint32{}, fmt.Errorf("transposed matrix is invalid: %w", err)
	}

	return t, nil
}

func (m cCSMatrix) Convert() (cRSMatrix, error) {
	return transpose(m)
}

func (m cRSMatrix) Convert() (cCSMatrix, error) {
	return transpose(m)
}

func multiply(a []uint32, x []float64, result []float64) {
	// major as in row-major order or column-major order
	majorPos := 0
	result[majorPos] = 0
	for i := 0; i < len(a); i++ {
		if a[i] != sen {
			result[majorPos] += x[a[i]]
		} else {
			majorPos++
			if majorPos < len(result) {
				result[majorPos] = 0
			}
		}
	}
}

// Calculates a matrix-vector multiply:
// result = a*x where a is a is m-by-n matrix and x is column vector with length n
func (a cRSMatrix) MatrixVectorMultiply(x []float64, result []float64) {
	multiply(a, x, result)
}

// Calculates a vector-matrix multiple:
// result = x*a where x is row vector with length m and a is a is m-by-n matrix
func (a cCSMatrix) VectorMatrixMultiply(x []float64, result []float64) {
	multiply(a, x, result)
}

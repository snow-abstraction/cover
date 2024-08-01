/*
 Copyright (C) 2024 Douglas Wayne Potter

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

package solvers

import "log"

// Calculate a lower bound for set covering problem instance specified by
// the binary matrix aC where element aC_{ij} = 1 iff element i is in subset j.
//
// This is done using a subgradient algorithm
// on the Lagrangian dual of the integer linear programming formulation of
// the set covering problem.
// TODO: pass in transpose instead of re-calculating on every call/
func CalcScLb(aC cCSMatrix /* C for column storage*/, costs []float64) (float64, error) {
	var n_cols int
	for i := 0; i < len(aC); i++ {
		if aC[i] == sen {
			n_cols++
		}
	}

	aR, err := aC.Convert() // R for row storage
	if err != nil {
		return 0, err
	}

	var n_rows int
	for i := 0; i < len(aR); i++ {
		if aR[i] == sen {
			n_rows++
		}
	}

	// The primal column vector
	x := make([]float64, n_cols)

	// The dual row vector commonly denoted by μ (the Greek "my")
	// u >= 0
	u := make([]float64, n_rows)

	// for storing the result of u*a
	uA := make([]float64, n_cols)

	// for storing results of a*x
	ax := make([]float64, n_rows)

	// max iterations
	n := 1000

	step := 0.001
	for k := 0; k < n; k++ {
		// We calc "1. update u" and then "2. find x" since then the values
		// are useable after the loop to calculate the upper bound
		// i.e. the Lagrangian Dual objective value

		// 1. update u: given x, take step following the subgradient (1 - Ax)
		row := 0
		// TODO: this special handling of the first row/col outside (before and after)
		/// the loop repeats code. See multiple function as well.
		// One idea is to use an initial sentinel and final sentinel,
		// with no indices allowed before the initial sentinel and no indices allowed
		for j := 0; j < len(aR); j++ {
			if aR[j] != sen {
				u[row] -= step * x[aR[j]] // one nonzero component of Ax from (1 - Ax)
			} else {
				u[row] += step // add step*1 where the 1 is the 1 in (1 - Ax) for the row
				// project u
				if u[row] < 0 {
					u[row] = 0
				}

				row++
			}
		}

		// 2. find x: given u
		// that is set x_i such that it is minimizes:
		// c(x) + u(1 - Ax) = (c - uA)x + u*1
		aC.VectorMatrixMultiply(u, uA)
		for i := 0; i < n_cols; i++ {
			if uA[i] <= costs[i] {
				x[i] = 0
			} else {
				x[i] = 1
			}
		}

		objectiveValue := calcObjectiveValue(n_cols, costs, x, aR, ax, n_rows, u)
		if k%100 == 0 {
			log.Printf("Objective value: %f", objectiveValue)
		}

		// TODO: use a better step length rule
		step = 1.0 / (1.0 + float64(k))
	}

	// calculate the lower bound i.e. the Lagrangian Dual objective value
	objectiveValue := calcObjectiveValue(n_cols, costs, x, aR, ax, n_rows, u)

	return objectiveValue, nil

}

func calcObjectiveValue(n_cols int, costs []float64, x []float64, aR cRSMatrix, ax []float64, n_rows int, u []float64) float64 {
	objectiveValue := 0.0

	for i := 0; i < n_cols; i++ {
		objectiveValue += costs[i] * x[i]
	}

	aR.MatrixVectorMultiply(x, ax)
	for j := 0; j < n_rows; j++ {
		objectiveValue += (u[j] * (1 - ax[j]))
	}
	return objectiveValue
}

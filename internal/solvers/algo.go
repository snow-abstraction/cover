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

// Calculate a lower bound for the (non-exact) set covering problem instance specified by
// the binary matrix aC where element aC_{ij} = 1 iff element i is in subset j.
//
// This is done using a subgradient algorithm
// on the Lagrangian Dual of the ILP (Integer Linear Programming) formulation of
// the set covering problem.
//
// The following is a reminder of the optimization mathematics.
// First let:
// X := {x | x_i \in {0, 1}, \all i}
// A := Ac
//
// Then this is the ILP formulation of the set covering problem.
// min_{x \in X} cx
// s.t. Ax >= 1
// X := {x | x_i \in {0, 1}, \all i}
// A := Ac
//
// And this is the Lagrangian Dual:
// max_{u >= 0} (min_{x } cx + u(1 - Ax))
//
// TODO: pass in transpose instead of re-calculating on every call
func CalcScLb(aC cCSMatrix /* C for column storage*/, costs []float64) (float64, error) {
	var nCols int
	for i := 0; i < len(aC); i++ {
		if aC[i] == sen {
			nCols++
		}
	}

	aR, err := aC.Convert() // R for row storage
	if err != nil {
		return 0, err
	}

	var nRows int
	for i := 0; i < len(aR); i++ {
		if aR[i] == sen {
			nRows++
		}
	}

	// TODO: initialized smart. User instance or previous node
	initialStepLength := 1.0

	// The primal column vector
	x := make([]float64, nCols)

	// The dual row vector commonly denoted by μ (the Greek "my")
	// u >= 0
	u := make([]float64, nRows)

	// for storing the result of u*aC
	uaC := make([]float64, nCols)

	// for storing results of aR*x
	aRx := make([]float64, nRows)

	// max iterations
	n := 1000
	nextLogK := 1

	for k := 0; k < n; k++ {
		// TODO: use a better step length rule
		step := initialStepLength / (1.0 + float64(k))

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
		aC.VectorMatrixMultiply(u, uaC)
		for i := 0; i < nCols; i++ {
			if uaC[i] <= costs[i] {
				x[i] = 0
			} else {
				x[i] = 1
			}
		}

		if k > nextLogK {
			nextLogK *= 2
			objectiveValue := calcObjectiveValue(nCols, costs, x, aR, aRx, nRows, u)
			log.Printf("Iteration %d Objective value: %f", k, objectiveValue)
		}
	}

	// calculate the lower bound i.e. the Lagrangian Dual objective value
	objectiveValue := calcObjectiveValue(nCols, costs, x, aR, aRx, nRows, u)

	return objectiveValue, nil

}

func calcObjectiveValue(n_cols int, costs []float64, x []float64, aR cRSMatrix, aRx []float64,
	nRows int, u []float64) float64 {
	var objectiveValue float64
	for i := 0; i < n_cols; i++ {
		objectiveValue += costs[i] * x[i]
	}

	aR.MatrixVectorMultiply(x, aRx)
	for j := 0; j < nRows; j++ {
		objectiveValue += (u[j] * (1 - aRx[j]))
	}
	return objectiveValue
}

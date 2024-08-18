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

type lagrangianDualResult struct {
	dualObjectiveValue float64
	// Subset indices of the cover. It may be be infeasible.
	primalSolution []int
	// If the primalSolution is proven to be an optimal solution to the exact
	// set cover problem.
	provenOptimalExact bool
	// Index of element not covered exactly. -1 if all covered exactly.
	notCoveredExactly int
}

// Calculate a lower bound for the (non-exact) set covering problem instance specified by
// the binary matrix aC where element aC_{ij} = 1 iff element i is in subset j.
//
// This is done using a subgradient algorithm
// on the Lagrangian Dual of the ILP (Integer Linear Programming) formulation of
// the set covering problem.
func CalcScLb(aC cCSMatrix /* C for column storage*/, costs []float64) (float64, error) {
	result, err := runDualIterations(aC, costs)
	if err != nil {
		return 0, err
	}
	return result.dualObjectiveValue, nil

}

// Run a subgradient algorithm
// on the Lagrangian Dual of the ILP (Integer Linear Programming) formulation of
// the set covering problem instance specified by
// the binary matrix aC where element aC_{ij} = 1 iff element i is in subset j.
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
func runDualIterations(aC cCSMatrix /* C for column storage*/, costs []float64) (lagrangianDualResult, error) {

	var nCols int
	for i := 0; i < len(aC); i++ {
		if aC[i] == sen {
			nCols++
		}
	}

	aR, err := aC.Convert() // R for row storage
	if err != nil {
		return lagrangianDualResult{}, err
	}

	var nRows int
	for i := 0; i < len(aR); i++ {
		if aR[i] == sen {
			nRows++
		}
	}

	// TODO: initialize step length smarter. Scale for costs.
	// Use instance or previous node. Set a value such some rows will be
	// cover after a few iterations.
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
	nextLogK := 100

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
				// TODO: think about overflow and precision issues here.
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

	return calcLagrangianDualResult(nCols, costs, x, aR, aRx, nRows, u), nil
}

func calcObjectiveValue(nCols int, costs []float64, x []float64, aR cRSMatrix, aRx []float64,
	nRows int, u []float64) float64 {
	result := calcLagrangianDualResult(nCols, costs, x, aR, aRx, nRows, u)
	return result.dualObjectiveValue
}

func calcLagrangianDualResult(nCols int, costs []float64, x []float64, aR cRSMatrix, aRx []float64,
	nRows int, u []float64) lagrangianDualResult {

	result := lagrangianDualResult{
		dualObjectiveValue: 0.0,
		primalSolution:     make([]int, 0, nRows),
		provenOptimalExact: true,
		notCoveredExactly:  -1,
	}

	for i := 0; i < nCols; i++ {
		result.dualObjectiveValue += costs[i] * x[i]
		if x[i] == 1.0 {
			result.primalSolution = append(result.primalSolution, i)
		}
	}

	aR.MatrixVectorMultiply(x, aRx)
	for j := 0; j < nRows; j++ {
		g := (1 - aRx[j])
		result.dualObjectiveValue += (u[j] * g)
		if g != 0.0 {
			result.provenOptimalExact = false
			result.notCoveredExactly = j
		}
		// if g_j == 0 then
		// 1. x is feasible w.r.t. g_j
		// 2. g_j*u_j == 0 so complementary slackness is fulfilled
		// 3. element/row is exactly covered
	}

	return result
}

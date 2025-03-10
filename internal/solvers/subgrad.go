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

import "log/slog"

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

	initialStepLength := calcMeanElementCost(aC, costs, nCols)

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
	nextCheckStatus := 1

	for k := 0; k < n; k++ {
		// TODO: use a better step length rule
		step := initialStepLength / (1.0 + float64(k))

		// We calc "1. update u" and then "2. find x" since then the values
		// are useable after the loop to calculate the upper bound
		// i.e. the Lagrangian Dual objective value

		// 1. update u: given x, take step following the subgradient (1 - Ax)
		row := 0
		isSubgradientZero := true
		aContrib := 0.0
		for _, colIdx := range aR {
			if colIdx != sen {
				if x[colIdx] == 1.0 {
					aContrib++
				}
				// TODO: think about overflow and precision issues here.
				//u[row] -= step * x[colIdx] // one nonzero component of Ax from (1 - Ax)
			} else {
				if aContrib != 1.0 {
					isSubgradientZero = false
				}
				u[row] += step * (1.0 - aContrib) // add step*1 where the 1 is the 1 in (1 - Ax) for the row
				aContrib = 0
				// project u
				if u[row] < 0 {
					u[row] = 0
				}

				row++
			}
		}

		if isSubgradientZero {
			result := calcLagrangianDualResult(nCols, costs, x, aR, aRx, nRows, u)
			slog.Debug("Stop iterating. Subgradient zero")
			return result, nil
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

		if k > nextCheckStatus {
			nextCheckStatus *= 2
			result := calcLagrangianDualResult(nCols, costs, x, aR, aRx, nRows, u)
			slog.Debug("Iteration status", "i", k, "objective value", result.dualObjectiveValue)
			if result.provenOptimalExact {
				slog.Debug("Stop iterating. Proven optimal")
				return result, nil
			}
		}
	}

	return calcLagrangianDualResult(nCols, costs, x, aR, aRx, nRows, u), nil
}

func calcMeanElementCost(aC cCSMatrix, costs []float64, nCols int) float64 {
	var meanElementCost float64
	var colIdx int
	var nnzInColumn int
	for _, rowIdx := range aC {
		if rowIdx == sen {
			meanElementCost += costs[colIdx] / (float64(nnzInColumn) * float64(nCols))
			nnzInColumn = 0
			colIdx++
		} else {
			nnzInColumn++
		}
	}
	return meanElementCost
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

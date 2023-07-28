/*
 Copyright (C) 2023 Douglas Wayne Potter

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

import (
	"math"
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

func TestBruteVersusBrute2(t *testing.T) {
	const maxElements = 6

	for n := 0; n < maxElements; n++ {
		maxSubsets := int(math.Pow(float64(2), float64(n)) - 1)
		maxM := maxSubsets / 2
		for m := 0; m < maxM; m++ {
			for seed := 0; seed < 10; seed++ {
				ins := cover.MakeRandomInstance(m, n, 1)
				ins2, err := MakeInstance(ins.N, ins.Subsets, ins.Costs)
				assert.NilError(t, err)
				sol, err := SolveByBruteForce(ins2)
				assert.NilError(t, err)
				sol2, err := SolveByBruteForce(ins2)
				assert.NilError(t, err)
				assert.DeepEqual(t, sol, sol2)
			}
		}
	}
}

func TestInfeasible2(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}}, []float64{1.0, 1.0, 1.0})
	assert.NilError(t, err)
	result, err := SolveByBruteForce2(ins)
	assert.NilError(t, err)
	assert.Assert(t, !result.ExactlyCovered, "should be infeasible")
}

func TestEmptyInstance2(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	result, err := SolveByBruteForce2(ins)
	assert.NilError(t, err)
	//  The result for an empty instance should be a feasible and itself be empty.
	emptyCover := subsetsEval{ExactlyCovered: true}
	assert.DeepEqual(t, result, emptyCover)
}

func TestCheaperSolutionFound2(t *testing.T) {
	ins, err := MakeInstance(2, [][]int{{0, 1}, {0}, {1}, {0}}, []float64{17, 7, 5, 3})
	assert.NilError(t, err)
	result, err := SolveByBruteForce2(ins)
	assert.NilError(t, err)
	theMinimum := subsetsEval{SubsetsIndices: []int{2, 3}, ExactlyCovered: true, Cost: 5 + 3}
	assert.DeepEqual(t, result, theMinimum)
}

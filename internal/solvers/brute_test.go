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

package solvers

import (
	"testing"

	"gotest.tools/v3/assert"
)

// Test when all possible sets of subsets result in either some elements not
// being covered or some elements being overcovered.
func TestInfeasible(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}}, []float64{1.0, 1.0, 1.0})
	assert.NilError(t, err)
	result, err := SolveByBruteForce(ins)
	assert.NilError(t, err)
	assert.Assert(t, !result.ExactlyCovered, "should be infeasible")
}

func TestEmptyInstance(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	result, err := SolveByBruteForce(ins)
	assert.NilError(t, err)
	//  The result for an empty instance should be a feasible and itself be empty.
	emptyCover := subsetsEval{ExactlyCovered: true}
	assert.DeepEqual(t, result, emptyCover)
}

func TestCheaperSolutionFound(t *testing.T) {
	ins, err := MakeInstance(2, [][]int{{0, 1}, {0}, {1}, {0}}, []float64{17, 7, 5, 3})
	assert.NilError(t, err)
	result, err := SolveByBruteForce(ins)
	assert.NilError(t, err)
	theMinimum := subsetsEval{SubsetsIndices: []int{2, 3}, ExactlyCovered: true, Cost: 5 + 3}
	assert.DeepEqual(t, result, theMinimum)
}

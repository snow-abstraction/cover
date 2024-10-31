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

// doctest package is for testing code used in documentation.
package doctest

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/solvers"
)

func TestReadMeExample(t *testing.T) {
	instance := cover.Instance{
		M:       4,
		Subsets: [][]int{{0}, {0, 1}, {1, 2}, {1}, {0, 1, 2, 3}, {2, 3}, {0, 1, 3}, {2}},
		Costs:   []float64{1.8, 1.7, 2.4, 1.4, 5.4, 2.7, 1.9, 1.6}}

	result, err := solvers.SolveByBranchAndBound(instance)
	assert.NilError(t, err)
	assert.Assert(t, result.Optimal)
	assert.Equal(t, result.Cost, 3.5)
	assert.DeepEqual(t, result.SubsetsIndices, []int{6, 7})
}

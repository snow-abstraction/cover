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

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func TestMakeInstanceWithDuplicateSubsets(t *testing.T) {
	_, err := MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}, {1, 2}}, []float64{1.0, 1.0, 1.0, 1.0})
	assert.NilError(t, err)
}

func TestMakeInstanceSubsetsNotReOrdered(t *testing.T) {
	subsets := [][]int{{0, 1}, {1, 2}, {0, 2}}
	ins, err := MakeInstance(3, subsets, []float64{1.0, 1.0, 1.0})
	assert.NilError(t, err)
	assert.DeepEqual(t, ins.subsets, [][]int{{0, 1}, {1, 2}, {0, 2}})
	slices.SortFunc(subsets, slices.Compare)
	assert.DeepEqual(t, subsets, [][]int{{0, 1}, {0, 2}, {1, 2}})
}

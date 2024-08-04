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

	"gonum.org/v1/gonum/stat/combin"
	"gotest.tools/v3/assert"
)

func TestMakeInstanceWithDuplicateSubsets(t *testing.T) {
	_, err := MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}, {1, 2}}, []float64{1.0, 1.0, 1.0, 1.0})
	assert.Assert(t, err != nil)
}

func TestMakeInstanceSubsetsNotReOrdered(t *testing.T) {
	subsets := [][]int{{0, 1}, {1, 2}, {0, 2}}
	ins, err := MakeInstance(3, subsets, []float64{1.0, 1.0, 1.0})
	assert.NilError(t, err)
	assert.DeepEqual(t, ins.subsets, [][]int{{0, 1}, {1, 2}, {0, 2}})
	slices.SortFunc(subsets, slices.Compare)
	assert.DeepEqual(t, subsets, [][]int{{0, 1}, {0, 2}, {1, 2}})
}

func BenchmarkCheckSubsetsForDuplicates(b *testing.B) {
	// The combinations of the first 23 elements are use to get unique subsets.
	upperBoundElement := 23
	subsets := combin.Combinations(upperBoundElement, 10)
	assert.Assert(b, len(subsets) == 1144066) // expected number of combinations

	// Use moreElements to add 100 identical elements to each subset for more
	// memory traffic. The subsets remain unique from the initial elements from
	// the combinations.
	//
	// Because the sorting used by checkSubsetsForDuplicates presumably
	// starts comparing subsets at the beginning these subsets are faster to check than
	// if they had started with identical elements and ended with unique elements.
	moreElements := make([]int, 100) // 100 because I don't have much free RAM
	for i := 0; i < len(moreElements); i++ {
		moreElements[i] = i + upperBoundElement
	}

	for i := 0; i < len(subsets); i++ {
		subsets[i] = append(subsets[i], moreElements...)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := checkSubsetsForDuplicates(subsets)
		assert.NilError(b, err)
	}
}

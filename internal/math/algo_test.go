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
package math

import (
	"math/rand"
	"testing"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/solvers"
	"gotest.tools/v3/assert"
)

func Test2(t *testing.T) {
	m := 6
	n := 30
	for i := 0; i < 100; i++ {
		seed := int64(rand.Int63())
		//seed := int64(944779272699051336)
		t.Logf("seed %d", seed)
		ins := cover.MakeRandomInstance(m, n, seed)
		solverInstance, err := solvers.MakeInstance(ins.M, ins.Subsets, ins.Costs)
		assert.NilError(t, err)

		// subsets := cCSMatrix{0, 1, sen, 1, 2, sen, 0, 1, sen}

		matrixData := make([]uint32, len(solverInstance.Subsets))
		for _, subset := range solverInstance.Subsets {
			for _, setIndex := range subset {
				matrixData = append(matrixData, uint32(setIndex))
			}
			matrixData = append(matrixData, sen)

		}
		matrix, err := makeCompressedColumnMatrix(matrixData)
		assert.NilError(t, err)
		result, err := solvers.SolveByBruteForce(solverInstance)
		assert.NilError(t, err)

		lb, err := CalcScLb(matrix, solverInstance.Costs)
		assert.NilError(t, err)
		t.Logf("%f %f", lb, result.Cost)
		assert.Check(t, lb <= result.Cost)
		// t.Fatalf("%f %f", lb, result.Cost)
		// assert.NilError(t, err)
	}
}

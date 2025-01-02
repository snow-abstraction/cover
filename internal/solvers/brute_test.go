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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

// Test when all possible sets of subsets result in either some elements not
// being covered or some elements being overcovered.
func TestInfeasible(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}}, []float64{1.0, 1.0, 1.0})
	assert.NilError(t, err)
	result, err := SolveByBruteForceInternal(ins)
	assert.NilError(t, err)
	assert.Assert(t, !result.ExactlyCovered, "should be infeasible")
}

func TestEmptyInstance(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	result, err := SolveByBruteForceInternal(ins)
	assert.NilError(t, err)
	//  The result for an empty instance should be a feasible and itself be empty.
	emptyCover := subsetsEval{ExactlyCovered: true, Optimal: true}
	assert.DeepEqual(t, result, emptyCover)
}

func TestCheaperSolutionFound(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1, 2}, {0}, {1}, {1, 2}, {0, 2}}, []float64{17, 5, 4, 3, 3})
	assert.NilError(t, err)
	result, err := SolveByBruteForceInternal(ins)
	assert.NilError(t, err)
	theMinimum := subsetsEval{SubsetsIndices: []int{2, 4}, ExactlyCovered: true, Cost: 7, Optimal: true}
	assert.DeepEqual(t, result, theMinimum)
}

func testBruteFindsEquallyGoodSolution(t *testing.T, spec cover.TestInstanceSpecification) {
	pythonResultBytes, err := os.ReadFile(filepath.Join("../..", spec.PythonSolutionPath))
	assert.NilError(t, err)
	var pythonResult map[string]interface{}
	err = json.Unmarshal(pythonResultBytes, &pythonResult)
	assert.NilError(t, err)

	instancePath := filepath.Join("../..", spec.InstancePath)
	ins, err := cover.ReadJsonInstance(instancePath)
	assert.NilError(t, err)

	result, err := SolveByBruteForce(*ins)
	assert.NilError(t, err)

	// This is tightly coupled to JSON format of tools/solve_sc.py.
	if result.ExactlyCovered {
		assert.Equal(t, "optimal", pythonResult["status"].(string))
		pythonCost := pythonResult["cost"].(float64)
		costDiff := math.Abs(result.Cost - pythonCost)
		assert.Assert(
			t,
			costDiff < 0.000000000001,
			"brute found an optimal cost %f but the Python script found an optimal cost %f ",
			result.Cost,
			pythonCost,
		)
	} else {
		assert.Equal(
			t,
			"infeasible",
			pythonResult["status"].(string),
			"was found infeasible by brute but not by the Python script",
		)
	}

}

func TestBruteOnTinyInstances(t *testing.T) {
	t.Parallel()
	instanceSpecifications := loadTinyInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBruteFindsEquallyGoodSolution(t, spec)
		})

	}
}

func BenchmarkBruteOnRandomTinyInstances(b *testing.B) {
	instanceSpecifications := loadTinyInstanceSpecifications(b)
	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		solverInstance := loadSolverInstance(b, filepath.Join("../..", spec.InstancePath))
		instances = append(instances, solverInstance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBruteForceInternal(instances[j])
			assert.NilError(b, err)
		}
	}
}

func TestComb5_3(t *testing.T) {
	expected := [][]int{
		{0, 1, 2},
		{0, 1, 3},
		{0, 1, 4},
		{0, 2, 3},
		{0, 2, 4},
		{0, 3, 4},
		{1, 2, 3},
		{1, 2, 4},
		{1, 3, 4},
		{2, 3, 4}}

	actual := make([][]int, 0, len(expected))
	generator := newCombinationGenerator(5, 3)

	for generator.next() {
		comb := make([]int, len(generator.combination))
		copy(comb, generator.combination)
		actual = append(actual, comb)
	}
	assert.DeepEqual(t, expected, actual)
}

func TestComb4_4(t *testing.T) {
	expected := [][]int{{0, 1, 2, 3}}

	actual := make([][]int, 0, len(expected))
	generator := newCombinationGenerator(4, 4)

	for generator.next() {
		comb := make([]int, len(generator.combination))
		copy(comb, generator.combination)
		actual = append(actual, comb)
	}
	assert.DeepEqual(t, expected, actual)
}

func TestComb0_0(t *testing.T) {
	c := newCombinationGenerator(0, 0)
	assert.Assert(t, !c.next())
}

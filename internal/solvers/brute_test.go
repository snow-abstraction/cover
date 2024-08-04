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
	ins, err := MakeInstance(3, [][]int{{0, 1, 2}, {0}, {1}, {1, 2}, {0, 2}}, []float64{17, 5, 4, 3, 3})
	assert.NilError(t, err)
	result, err := SolveByBruteForce(ins)
	assert.NilError(t, err)
	theMinimum := subsetsEval{SubsetsIndices: []int{2, 4}, ExactlyCovered: true, Cost: 7}
	assert.DeepEqual(t, result, theMinimum)
}

func testBruteFindsEquallyGoodSolution(t *testing.T, spec cover.TestInstanceSpecification) {

	pythonResultBytes, err := os.ReadFile(filepath.Join("../..", spec.PythonSolutionPath))
	assert.NilError(t, err)
	var pythonResult map[string]interface{}
	err = json.Unmarshal(pythonResultBytes, &pythonResult)
	assert.NilError(t, err)

	instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
	assert.NilError(t, err)
	var ins cover.Instance
	err = json.Unmarshal(instanceBytes, &ins)
	assert.NilError(t, err)
	solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
	assert.NilError(t, err)

	result, err := SolveByBruteForce(solverInstance)
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

func TestInstances(t *testing.T) {
	instanceSpecifications := loadInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBruteFindsEquallyGoodSolution(t, spec)
		})

	}
}

func BenchmarkRandomInstances(b *testing.B) {
	instanceSpecifications := loadInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
		assert.NilError(b, err)
		var ins cover.Instance
		err = json.Unmarshal(instanceBytes, &ins)
		assert.NilError(b, err)
		solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
		instances = append(instances, solverInstance)
		assert.NilError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBruteForce(instances[j])
			assert.NilError(b, err)
		}
	}

}

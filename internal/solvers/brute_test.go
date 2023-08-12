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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

const solveSCPATH = "../../tools/solve_sc.py"

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

func solveWithPythonScript(t *testing.T, ins cover.Instance) map[string]interface{} {
	// Save instance to a temporary file
	f, err := os.CreateTemp("", "test_instance")
	assert.NilError(t, err)
	defer os.Remove(f.Name())
	b, err := json.MarshalIndent(ins, "", "  ")
	assert.NilError(t, err)
	err = os.WriteFile(f.Name(), b, 0600)
	assert.NilError(t, err)

	// Solve instance using the python script
	cmd := exec.Command("python", solveSCPATH, f.Name())
	stdout, err := cmd.Output()
	assert.NilError(t, err)

	// Extract the result from stdout
	s := string(stdout)
	const resultDelimiter = "solve_sc_result:"
	assert.Assert(t, strings.Contains(s, resultDelimiter))
	splitOutput := strings.Split(s, resultDelimiter)
	resultStr := splitOutput[len(splitOutput)-1]
	var result map[string]interface{}
	err = json.Unmarshal([]byte(resultStr), &result)
	assert.NilError(t, err)

	return result

}

func testBruteFindsEquallyGoodSolution(t *testing.T, m int, n int, seed int64) {
	instanceInfo := fmt.Sprintf(
		"for instance generated using [m: %d, n: %d, random seed: %d]",
		m,
		n,
		seed,
	)
	t.Log(instanceInfo)
	ins := cover.MakeRandomInstance(m, n, seed)

	pythonResult := solveWithPythonScript(t, ins)

	solverInstance, err := MakeInstance(ins.N, ins.Subsets, ins.Costs)
	assert.NilError(t, err)
	result, err := SolveByBruteForce(solverInstance)
	assert.NilError(t, err)

	// This is tightly coupled to JSON format of tools/solve_sc.py.
	if result.ExactlyCovered {
		assert.Equal(t, "optimal", pythonResult["status"].(string), instanceInfo)
		pythonCost := pythonResult["cost"].(float64)
		costDiff := math.Abs(result.Cost - pythonCost)
		assert.Assert(
			t,
			costDiff < 0.000000000001,
			"%s, brute found an optimal cost %f but %s found an optimal cost %f ",
			instanceInfo,
			result.Cost,
			solveSCPATH,
			pythonCost,
		)
	} else {
		assert.Equal(
			t,
			"infeasible",
			pythonResult["status"].(string),
			"%s was found infeasible by brute but not by %s",
			instanceInfo,
			solveSCPATH,
		)
	}

}

func TestRandomInstances(t *testing.T) {
	seed := int64(1) // random seed
	maxN := 8        // max number of elements

	for n := 1; n < maxN; n++ {
		maxM := int(math.Exp2(float64(n))) / 2
		for m := 1; m <= maxM; m *= 2 {
			testBruteFindsEquallyGoodSolution(t, m, n, seed)
			seed++
		}
	}
}

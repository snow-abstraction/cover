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

// bb = branch-and-bound

package solvers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/tree"
	"gotest.tools/v3/assert"
)

func TestRemoveMoreExpensiveDuplicatesTrivial(t *testing.T) {
	input := instance{1, [][]int{{0}, {0}, {0}}, []float64{3, 2, 1}}
	output, indices := removeMoreExpensiveDuplicates(input)
	assert.DeepEqual(t, instance{1, [][]int{{0}}, []float64{1}}, output, cmp.AllowUnexported(instance{}))
	assert.DeepEqual(t, []int{2}, indices)
}

func TestRemoveMoreExpensiveDuplicatesSmall(t *testing.T) {
	input := instance{2, [][]int{{0, 1}, {0}, {1}, {1}, {0}, {0, 1}}, []float64{13, 11, 7, 5, 3, 2}}
	output, indices := removeMoreExpensiveDuplicates(input)
	assert.DeepEqual(t, instance{2, [][]int{{0}, {0, 1}, {1}}, []float64{3, 2, 5}}, output, cmp.AllowUnexported(instance{}))
	assert.DeepEqual(t, []int{4, 5, 3}, indices)
}

func TestCreateSubInstanceOnEmptyInstance(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	subproblemIns, err := createSubInstance(ins, tree.CreateRoot())
	assert.NilError(t, err)
	assert.DeepEqual(t, subproblemIns.ins, ins, cmp.AllowUnexported(instance{}))
}

func TestCreateSubInstanceCreatesInfeasibleBranch(t *testing.T) {
	ins, err := MakeInstance(2, [][]int{{0, 1}, {0}}, []float64{1, 2})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	_, diffNode := rootNode.Branch(0, 0, 1)

	actualDiffIns, err := createSubInstance(ins, diffNode)
	assert.NilError(t, err)
	assert.Assert(t, actualDiffIns == nil)
}

func TestCreateSubInstancesSimple(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {0}, {1}, {2}}, []float64{1, 2, 3, 4})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	bothNode, diffNode := rootNode.Branch(0, 0, 1)

	actualBothIns, err := createSubInstance(ins, bothNode)
	assert.NilError(t, err)
	expectedBothIns, err := MakeInstance(3, [][]int{{0, 1}, {2}}, []float64{1, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBothIns, actualBothIns.ins, cmp.AllowUnexported(instance{}))

	actualDiffIns, err := createSubInstance(ins, diffNode)
	assert.NilError(t, err)
	expectedDiffIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}}, []float64{2, 3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiffIns, actualDiffIns.ins, cmp.AllowUnexported(instance{}))
}

func TestCreateSubInstancesTricker(t *testing.T) {
	ins, err := MakeInstance(
		3,
		[][]int{{0}, {1}, {2}, {0, 1}, {0, 2}, {1, 2}, {0, 1, 2}},
		[]float64{1, 2, 3, 4, 5, 6, 7})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	both01Node, diff01Node := rootNode.Branch(0, 0, 1)
	both01Both12Node, both01Diff12Node := both01Node.Branch(0, 1, 2)
	diff01both12Node, diff01Diff12Node := diff01Node.Branch(0, 1, 2)

	actualBoth01Both12NodeIns, err := createSubInstance(ins, both01Both12Node)
	assert.NilError(t, err)
	expectedBoth01Both12NodeIns, err := MakeInstance(3, [][]int{{0, 1, 2}}, []float64{7})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Both12NodeIns, actualBoth01Both12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualBoth01Diff12NodeIns, err := createSubInstance(ins, both01Diff12Node)
	assert.NilError(t, err)
	expectedBoth01Diff12NodeIns, err := MakeInstance(3, [][]int{{2}, {0, 1}}, []float64{3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Diff12NodeIns, actualBoth01Diff12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualDiff01bBoth12NodeIns, err := createSubInstance(ins, diff01both12Node)
	assert.NilError(t, err)
	expectedDiff01Both12NodeIns, err := MakeInstance(3, [][]int{{0}, {1, 2}}, []float64{1, 6})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Both12NodeIns, actualDiff01bBoth12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualDiff01Diff12NodeIns, err := createSubInstance(ins, diff01Diff12Node)
	assert.NilError(t, err)
	expectedDiff01Diff12NodeIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}, {0, 2}}, []float64{1, 2, 3, 5})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Diff12NodeIns, actualDiff01Diff12NodeIns.ins, cmp.AllowUnexported(instance{}))
}

func testBBFindsEquallyGoodSolution(t *testing.T, spec cover.TestInstanceSpecification) {
	pythonResultBytes, err := os.ReadFile(filepath.Join("../..", spec.PythonSolutionPath))
	assert.NilError(t, err)
	var pythonResult map[string]interface{}
	err = json.Unmarshal(pythonResultBytes, &pythonResult)
	assert.NilError(t, err)
	solverInstance := loadSolverInstance(t, filepath.Join("../..", spec.InstancePath))
	result, err := SolveByBranchAndBoundInternal(solverInstance, createTestBranchAndBoundConfig())
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

func TestBBOnTinyInstances(t *testing.T) {
	t.Parallel()
	instanceSpecifications := loadTinyInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBBFindsEquallyGoodSolution(t, spec)
		})
	}
}

func TestBBOnSmallInstances(t *testing.T) {
	t.Parallel()
	instanceSpecifications := loadSmallInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBBFindsEquallyGoodSolution(t, spec)
		})

	}
}

func BenchmarkBBOnRandomTinyInstances(b *testing.B) {
	instanceSpecifications := loadTinyInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		solverInstance := loadSolverInstance(b, filepath.Join("../..", spec.InstancePath))
		instances = append(instances, solverInstance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBoundInternal(instances[j], createTestBranchAndBoundConfig())
			assert.NilError(b, err)
		}
	}
}

func BenchmarkBBOnRandomScale1TinyInstances(b *testing.B) {
	instanceSpecifications := loadTinyInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		if spec.CostScale != 1.0 {
			continue
		}
		solverInstance := loadSolverInstance(b, filepath.Join("../..", spec.InstancePath))
		instances = append(instances, solverInstance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBoundInternal(instances[j], createTestBranchAndBoundConfig())
			assert.NilError(b, err)
		}
	}
}

func BenchmarkBBOnRandomScale1000TinyInstances(b *testing.B) {
	instanceSpecifications := loadTinyInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		if spec.CostScale != 1000.0 {
			continue
		}
		solverInstance := loadSolverInstance(b, filepath.Join("../..", spec.InstancePath))
		instances = append(instances, solverInstance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBoundInternal(instances[j], createTestBranchAndBoundConfig())
			assert.NilError(b, err)
		}
	}
}

func BenchmarkBBOnRandomSmallInstances(b *testing.B) {
	instanceSpecifications := loadSmallInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		solverInstance := loadSolverInstance(b, filepath.Join("../..", spec.InstancePath))
		instances = append(instances, solverInstance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBoundInternal(instances[j], createTestBranchAndBoundConfig())
			assert.NilError(b, err)
		}
	}
}
